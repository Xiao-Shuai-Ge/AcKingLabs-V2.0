package response

import (
	"errors"
	"fmt"
	"testing"
)

func TestMessage(t *testing.T) {
	if m := Message(CodeOK); m != "成功" {
		t.Errorf("Message(CodeOK) = %q", m)
	}
	if m := Message(CodeRateLimited); m == "未知错误" {
		t.Error("已定义错误码不应返回未知错误")
	}
	if m := Message(999999); m != "未知错误" {
		t.Errorf("未定义错误码应返回未知错误, got %q", m)
	}
}

func TestAppErr(t *testing.T) {
	err := NewErr(CodeNotFound)
	ae := AsAppErr(err)
	if ae == nil || ae.Code != CodeNotFound {
		t.Fatalf("AsAppErr 未识别业务错误")
	}
	if ae.Error() != Message(CodeNotFound) {
		t.Errorf("默认文案错误: %q", ae.Error())
	}

	custom := NewErrMsg(CodeBadRequest, "自定义文案")
	if custom.Error() != "自定义文案" {
		t.Errorf("自定义文案错误: %q", custom.Error())
	}

	if AsAppErr(errors.New("plain")) != nil {
		t.Error("普通 error 不应被识别为业务错误")
	}
	if AsAppErr(fmt.Errorf("wrapped: %w", NewErr(CodeForbidden))) == nil {
		t.Error("包装过的业务错误应能被识别")
	}
}
