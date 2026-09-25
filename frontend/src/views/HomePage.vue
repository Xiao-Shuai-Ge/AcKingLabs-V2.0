<script setup lang="ts">
// 主页：logo + 标题 + 一句话，入场动画编排：
// 依次是 logo 模糊浮现 -> 标题逐字上浮 -> 分隔线展开 -> 文案字距收拢；
// 背景两团柔光缓慢漂移，logo 带轻微鼠标视差与呼吸浮动。
import { onMounted, ref } from 'vue'

const started = ref(false)
const titleChars = 'AcKing 算法竞赛实验室'.split('')

onMounted(() => {
  // 不用 requestAnimationFrame：后台/未渲染的标签页里它不会触发，动画会被卡住
  setTimeout(() => {
    started.value = true
  }, 80)
})

// 轻量鼠标视差
const px = ref(0)
const py = ref(0)
function onMove(e: MouseEvent) {
  const cx = window.innerWidth / 2
  const cy = window.innerHeight / 2
  px.value = Math.max(-1, Math.min(1, (e.clientX - cx) / cx))
  py.value = Math.max(-1, Math.min(1, (e.clientY - cy) / cy))
}
</script>

<template>
  <div
    class="relative min-h-screen flex flex-col items-center justify-center px-4 overflow-hidden"
    @mousemove="onMove"
  >
    <!-- 背景柔光（缓慢漂移） -->
    <div class="hero-blob hero-blob-a" />
    <div class="hero-blob hero-blob-b" />

    <!-- Logo：视差 + 呼吸浮动 + 模糊浮现 -->
    <div
      class="hero-parallax z-10"
      :style="{ transform: `translate(${px * -10}px, ${py * -8}px)` }"
    >
      <div class="hero-float">
        <img
          src="/assets/AcKing_black.png"
          alt="实验室标志"
          class="hero-logo"
          :class="{ start: started }"
        />
      </div>
    </div>

    <!-- 标题：逐字上浮 -->
    <h1 class="z-10 mt-8 text-3xl md:text-4xl font-bold text-gray-800 select-none">
      <span
        v-for="(ch, i) in titleChars"
        :key="i"
        class="hero-char inline-block"
        :class="{ start: started }"
        :style="{ animationDelay: `${0.4 + i * 0.05}s` }"
        >{{ ch === ' ' ? ' ' : ch }}</span
      >
    </h1>

    <!-- 分隔线：由中心展开 -->
    <div class="hero-divider z-10" :class="{ start: started }" />

    <!-- 文案：字距收拢浮现 -->
    <p class="hero-tagline z-10 text-gray-600 text-center leading-relaxed text-base md:text-xl">
      我们致力于为校内算法竞赛爱好者提供一个良好的学习交流环境
    </p>
  </div>
</template>

<style scoped>
/* Logo 入场：模糊 + 缩放 -> 清晰 */
.hero-logo {
  width: min(60vw, 22rem);
  opacity: 0;
  transform: scale(0.92);
  filter: blur(14px);
}
.hero-logo.start {
  animation: logo-in 1.1s cubic-bezier(0.22, 1, 0.36, 1) 0.05s forwards;
}
@keyframes logo-in {
  to {
    opacity: 1;
    transform: scale(1);
    filter: blur(0);
  }
}

/* Logo 呼吸浮动 */
.hero-float {
  animation: float 6s ease-in-out infinite alternate;
}
@keyframes float {
  from {
    transform: translateY(-6px);
  }
  to {
    transform: translateY(6px);
  }
}

.hero-parallax {
  transition: transform 0.4s cubic-bezier(0.22, 1, 0.36, 1);
  will-change: transform;
}

/* 标题逐字 */
.hero-char {
  opacity: 0;
  transform: translateY(20px);
  filter: blur(4px);
}
.hero-char.start {
  animation: char-in 0.55s cubic-bezier(0.22, 1, 0.36, 1) forwards;
}
@keyframes char-in {
  to {
    opacity: 1;
    transform: translateY(0);
    filter: blur(0);
  }
}

/* 分隔线由中心展开 */
.hero-divider {
  width: 7rem;
  height: 1px;
  margin: 2rem 0 1.75rem;
  background: #d1d5db;
  transform: scaleX(0);
  opacity: 0;
}
.hero-divider.start {
  animation: divider-in 0.8s cubic-bezier(0.22, 1, 0.36, 1) 0.95s forwards;
}
@keyframes divider-in {
  to {
    transform: scaleX(1);
    opacity: 1;
  }
}

/* 文案：宽字距收拢浮现 */
.hero-tagline {
  max-width: 42rem;
  opacity: 0;
  letter-spacing: 0.35em;
  text-indent: 0.35em; /* 抵消末字符字距，保持视觉居中 */
  transform: translateY(10px);
}
.hero-tagline.start {
  animation: tagline-in 1.1s cubic-bezier(0.22, 1, 0.36, 1) 1.1s forwards;
}
@keyframes tagline-in {
  to {
    opacity: 1;
    letter-spacing: 0.06em;
    text-indent: 0.06em;
    transform: translateY(0);
  }
}

/* 背景柔光 */
.hero-blob {
  position: absolute;
  border-radius: 9999px;
  filter: blur(90px);
  pointer-events: none;
}
.hero-blob-a {
  width: 34rem;
  height: 34rem;
  top: -6rem;
  left: -4rem;
  background: rgba(191, 219, 254, 0.5);
  animation: drift-a 19s ease-in-out infinite alternate;
}
.hero-blob-b {
  width: 30rem;
  height: 30rem;
  bottom: -5rem;
  right: -3rem;
  background: rgba(233, 213, 255, 0.45);
  animation: drift-b 23s ease-in-out infinite alternate;
}
@keyframes drift-a {
  to {
    transform: translate(3rem, 2.5rem) scale(1.08);
  }
}
@keyframes drift-b {
  to {
    transform: translate(-2.5rem, -3rem) scale(1.1);
  }
}

/* 尊重系统减弱动效设置 */
@media (prefers-reduced-motion: reduce) {
  .hero-logo,
  .hero-char,
  .hero-divider,
  .hero-tagline {
    opacity: 1;
    transform: none;
    filter: none;
    animation: none;
    letter-spacing: normal;
    text-indent: 0;
  }
  .hero-float,
  .hero-blob-a,
  .hero-blob-b,
  .hero-parallax {
    animation: none;
  }
}
</style>
