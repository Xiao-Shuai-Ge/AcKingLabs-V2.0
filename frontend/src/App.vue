<script setup lang="ts">
import AppHeader from '@/components/AppHeader.vue'
import ToastStack from '@/components/ToastStack.vue'
</script>

<template>
  <AppHeader />
  <router-view />
  <ToastStack />
</template>
