<script setup lang="ts">
import { defineProps } from 'vue';

interface Car {
  id: string;
  make: string;
  model: string;
  license_plate: string;
  status: 'available' | 'rented' | 'inspecting' | 'maintenance';
  image_url?: string;
}

const props = defineProps<{
  car: Car;
}>();

const statusColors = {
  available: 'bg-green-100 text-green-800',
  rented: 'bg-red-100 text-red-800',
  inspecting: 'bg-yellow-100 text-yellow-800',
  maintenance: 'bg-gray-100 text-gray-800',
};

const statusLabels = {
  available: 'Available',
  rented: 'Rented',
  inspecting: 'Inspecting',
  maintenance: 'Maintenance',
};
</script>

<template>
  <div class="bg-white rounded-lg shadow-md overflow-hidden flex flex-col">
    <!-- Top: Car Image -->
    <div class="aspect-w-16 aspect-h-9 bg-gray-200">
      <img
        v-if="car.image_url"
        :src="car.image_url"
        :alt="`${car.make} ${car.model}`"
        class="w-full h-48 object-cover"
      />
      <div v-else class="w-full h-48 flex items-center justify-center text-gray-400">
        No Image
      </div>
    </div>

    <!-- Middle: Details -->
    <div class="p-4 flex-1">
      <div class="flex justify-between items-start">
        <div>
          <h3 class="text-xl font-bold text-gray-900">{{ car.make }} {{ car.model }}</h3>
          <p class="text-lg text-gray-600 font-mono mt-1">{{ car.license_plate }}</p>
        </div>
        <!-- Right Top: Status Badge -->
        <span
          class="px-3 py-1 rounded-full text-sm font-medium"
          :class="statusColors[car.status] || 'bg-gray-100 text-gray-800'"
        >
          {{ statusLabels[car.status] || car.status }}
        </span>
      </div>
    </div>

    <!-- Bottom: Action Button -->
    <div class="p-4 border-t border-gray-100">
      <button
        class="w-full h-12 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg shadow-sm transition-colors flex items-center justify-center text-lg touch-manipulation"
      >
        <span v-if="car.status === 'available'">Check Out</span>
        <span v-else-if="car.status === 'rented'">Start Inspection</span>
        <span v-else-if="car.status === 'inspecting'">Finish Inspection</span>
        <span v-else>Manage</span>
      </button>
    </div>
  </div>
</template>
