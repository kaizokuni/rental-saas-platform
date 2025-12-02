<script setup lang="ts">
defineProps<{
  title: string;
  value: string | number;
  icon?: string;
  trend?: string;
}>();
</script>

<template>
  <div class="bg-white overflow-hidden shadow rounded-lg">
    <div class="p-5">
      <div class="flex items-center">
        <div class="flex-shrink-0" v-if="icon">
          <!-- Icon placeholder -->
          <div class="h-6 w-6 text-gray-400">{{ icon }}</div>
        </div>
        <div class="ml-5 w-0 flex-1">
          <dl>
            <dt class="text-sm font-medium text-gray-500 truncate">
              {{ title }}
            </dt>
            <dd>
              <div class="text-lg font-medium text-gray-900">
                {{ value }}
              </div>
            </dd>
          </dl>
        </div>
      </div>
    </div>
    <div class="bg-gray-50 px-5 py-3" v-if="trend">
      <div class="text-sm">
        <span class="text-green-600 font-medium">
          {{ trend }}
        </span>
        <span class="text-gray-500">
          from last month
        </span>
      </div>
    </div>
  </div>
</template>
