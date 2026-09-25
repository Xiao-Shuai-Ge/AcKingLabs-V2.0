<script setup lang="ts">
import { teamMembers, missions } from '@/utils/team'
</script>

<template>
  <div class="pt-[60px] bg-gray-50 min-h-screen">
    <!-- 头部横幅 -->
    <div class="bg-gradient-to-r from-blue-500 to-purple-600 py-16">
      <div class="container mx-auto px-6 relative z-10 text-white">
        <img src="/assets/AcKing_white.png" alt="实验室标志" class="mx-auto w-32 h-auto" />
        <h1 class="text-center text-3xl md:text-4xl font-bold mt-4">AcKing 算法竞赛实验室</h1>
        <p class="text-center text-white/80 mt-3">
          我们致力于为校内算法竞赛爱好者提供一个良好的学习交流环境
        </p>
      </div>
    </div>

    <!-- 团队成员（横向大卡，内容上下居中） -->
    <div class="container mx-auto px-6 py-16">
      <div class="space-y-6">
        <div
          v-for="(m, i) in teamMembers"
          :key="m.username"
          class="bg-white rounded-lg shadow-md overflow-hidden transition-all duration-300 hover:shadow-xl hover:-translate-y-1 fade-in-up"
          :style="{ animationDelay: `${i * 0.1}s` }"
        >
          <div class="p-6 flex items-center gap-6 md:gap-8">
            <div
              class="w-24 h-24 md:w-32 md:h-32 shrink-0 rounded-full overflow-hidden transition-transform duration-300 hover:scale-105"
            >
              <img
                :src="m.avatar"
                :alt="m.username"
                class="w-full h-full object-cover object-top"
              />
            </div>
            <div class="flex-grow min-w-0">
              <h3 class="text-xl md:text-2xl font-bold text-gray-800 mb-2">
                {{ m.username }}
                <span class="text-base md:text-xl text-gray-500 font-normal ml-1">{{
                  m.role
                }}</span>
              </h3>
              <h4 class="text-md text-gray-600 mb-3">
                <i class="fa-solid fa-user mr-1" />
                {{ m.grade ? m.grade + ' ' : '' }}{{ m.name }}
              </h4>
              <p class="text-sm text-gray-500">{{ m.bio }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 关于我们（使命） -->
    <div class="bg-white py-16">
      <div class="container mx-auto px-6">
        <h2 class="text-3xl font-bold text-center mb-12 text-gray-800">关于我们</h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div
            v-for="mission in missions"
            :key="mission.title"
            class="p-6 rounded-lg text-center"
            :class="mission.card"
          >
            <div
              class="w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4"
              :class="mission.circle"
            >
              <i class="fa-solid text-white text-2xl" :class="mission.icon" />
            </div>
            <h3 class="text-xl font-bold mb-3 text-gray-800">{{ mission.title }}</h3>
            <p class="text-gray-600">{{ mission.desc }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 加入我们 -->
    <div class="bg-gray-50 py-16">
      <div class="container mx-auto px-6 text-center">
        <h2 class="text-3xl font-bold mb-6 text-gray-800">加入我们</h2>
        <p class="text-gray-600 max-w-2xl mx-auto mb-8">
          你可以通过大一的统一考核加入我们，也可以凭借过硬的实力破格加入我们，详情请联系我们。
        </p>
        <router-link
          to="/more/resume"
          class="inline-block px-10 py-3 rounded-lg bg-black text-white font-medium hover:bg-gray-800 transition-colors"
        >
          投递简历
        </router-link>
      </div>
    </div>
  </div>
</template>
