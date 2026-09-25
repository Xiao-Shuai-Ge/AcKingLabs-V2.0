package service

import (
	"bytes"
	"crypto/sha256"
	// 注册 webp 解码器（x/image/webp 仅支持解码）
	"encoding/hex"
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	xdraw "golang.org/x/image/draw"

	"acking/internal/app"
	"acking/pkg/response"
)

// 上传与图片处理参数
const (
	maxUploadBytes = 10 << 20  // 原始文件上限 10MB
	maxImageSide   = 1600      // 长边超过此值按比例缩小（站点展示足够）
	jpegQuality    = 82        // 常规编码质量
	jpegRetryQual  = 60        // 压缩后仍超预算时的降档质量（只降一次）
	jpegBudget     = 1500000   // JPEG 结果预算 1.5MB
	pngKeepBytes   = 300 << 10 // 无需处理的小 PNG 原样保留
)

// UploadImage 图片上传（登录后可用）。
// 存储通道：配置了 OSS 走 OSS（sha256 内容去重、公开读），否则本地磁盘。
// 压缩策略（尺寸优先，一次成型）：
//  1. 先按尺寸缩放——长边 > 1600px 用 CatmullRom 高质量重采样等比缩小
//     （手机照片一般 4000px+，缩到 1600px 后体积已大幅下降且观感无损）；
//  2. 再一次性编码 JPEG(q82)；仅当仍超 1.5MB 才降 q60 重编一次。
//  3. GIF（动图）与解码失败但体积合法的文件原样保存；带透明通道的
//     PNG/WebP 保留 PNG，超预算才平铺白底转 JPEG。
func UploadImage(file *multipart.FileHeader) (url string, err error) {
	if file.Size > maxUploadBytes {
		return "", response.NewErrMsg(response.CodeBadRequest, "图片大小不能超过 10MB")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif":
	default:
		return "", response.NewErrMsg(response.CodeBadRequest, "仅支持 png/jpg/jpeg/webp/gif 图片")
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	data, ext = processImage(data, ext)

	sum := sha256.Sum256(data)
	name := hex.EncodeToString(sum[:]) + ext
	key := time.Now().Format("200601") + "/" + name

	if app.Cfg.OSSConfigured() {
		return putOSS(key, data)
	}
	return putLocal(key, data)
}

// ---- 图片处理 ----

func decodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

func hasAlpha(img image.Image) bool {
	switch v := img.(type) {
	case *image.Gray, *image.CMYK, *image.YCbCr:
		return false
	case *image.NRGBA:
		// 逐像素扫 alpha（1600x1200 量级耗时毫秒级）
		for y := 0; y < v.Rect.Dy(); y++ {
			row := v.Pix[y*v.Stride : y*v.Stride+v.Rect.Dx()*4]
			for x := 3; x < len(row); x += 4 {
				if row[x] != 0xff {
					return true
				}
			}
		}
		return false
	default:
		b := img.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				if _, _, _, a := img.At(x, y).RGBA(); a != 0xffff {
					return true
				}
			}
		}
		return false
	}
}

// processImage 返回处理后的字节与最终扩展名
func processImage(data []byte, ext string) ([]byte, string) {
	// 动图原样保留
	if ext == ".gif" {
		return data, ext
	}
	img, err := decodeImage(data)
	if err != nil {
		// 解码失败（如动态 webp）：体积合法则原样保存
		return data, ext
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	// 带透明且尺寸已达标的小 PNG：无损保留
	alpha := hasAlpha(img)
	if alpha && ext == ".png" && len(data) <= pngKeepBytes && max(w, h) <= maxImageSide {
		return data, ext
	}

	// 长边超限：等比缩小（CatmullRom 高质量重采样）
	if max(w, h) > maxImageSide {
		scale := float64(maxImageSide) / float64(max(w, h))
		nw, nh := max(1, int(float64(w)*scale)), max(1, int(float64(h)*scale))
		dst := image.NewNRGBA(image.Rect(0, 0, nw, nh))
		xdraw.CatmullRom.Scale(dst, dst.Rect, img, b, xdraw.Src, nil)
		img = dst
		w, h = nw, nh
	}

	encodeJPEG := func(q int) []byte {
		var buf bytes.Buffer
		// JPEG 不支持透明：平铺白底
		flat := img
		if alpha {
			bg := image.NewNRGBA(image.Rect(0, 0, w, h))
			xdraw.Draw(bg, bg.Rect, image.White, image.Point{}, xdraw.Src)
			xdraw.Draw(bg, bg.Rect, img, image.Point{}, xdraw.Over)
			flat = bg
		}
		if err := jpeg.Encode(&buf, flat, &jpeg.Options{Quality: q}); err != nil {
			return data // 编码失败退回原文件
		}
		return buf.Bytes()
	}

	if alpha {
		// 透明图保留 PNG；超预算才平铺白底转 JPEG
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err == nil && buf.Len() <= 2*jpegBudget/3 {
			return buf.Bytes(), ".png"
		}
		out := encodeJPEG(jpegQuality)
		if len(out) > jpegBudget {
			out = encodeJPEG(jpegRetryQual)
		}
		return out, ".jpg"
	}

	out := encodeJPEG(jpegQuality)
	if len(out) > jpegBudget {
		out = encodeJPEG(jpegRetryQual)
	}
	return out, ".jpg"
}

// ---- 存储通道 ----

// putOSS 上传到阿里云 OSS：sha256 内容去重，公开读
func putOSS(key string, data []byte) (string, error) {
	cfg := app.Cfg.OSS
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return "", response.NewErrMsg(response.CodeInternal, "存储服务初始化失败")
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return "", response.NewErrMsg(response.CodeInternal, "存储桶不可用")
	}
	exists, err := bucket.IsObjectExist(key)
	if err == nil && exists {
		return objectURL(cfg.Endpoint, cfg.Bucket, key), nil
	}
	if err := bucket.PutObject(key, bytes.NewReader(data), oss.ObjectACL(oss.ACLPublicRead)); err != nil {
		return "", response.NewErrMsg(response.CodeInternal, "图片上传失败，请稍后重试")
	}
	return objectURL(cfg.Endpoint, cfg.Bucket, key), nil
}

func objectURL(endpoint, bucket, key string) string {
	return fmt.Sprintf("https://%s.%s/%s", bucket, endpoint, key)
}

// putLocal 本地磁盘存储（未配置 OSS 时的降级通道）
func putLocal(key string, data []byte) (string, error) {
	dir := filepath.Join(app.Cfg.App.UploadDir, filepath.Dir(key))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(app.Cfg.App.UploadDir, key)
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return "", err
		}
	}
	rel := "/uploads/" + key
	if app.Cfg.App.BaseURL != "" {
		rel = strings.TrimRight(app.Cfg.App.BaseURL, "/") + rel
	}
	return rel, nil
}
