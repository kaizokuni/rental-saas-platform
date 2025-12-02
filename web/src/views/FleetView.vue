<script setup lang="ts">
import { ref } from 'vue';
import { useQuery, useQueryClient } from '@tanstack/vue-query';
import { useAuthStore } from '../stores/auth';
import CarCard from '../components/fleet/CarCard.vue';
import AddCarModal from '../components/fleet/AddCarModal.vue';

const authStore = useAuthStore();
const queryClient = useQueryClient();
const isAddModalOpen = ref(false);

interface Car {
  id: string;
  make: string;
  model: string;
  license_plate: string;
  status: 'available' | 'rented' | 'inspecting' | 'maintenance';
  image_url?: string;
}

const fetchCars = async (): Promise<Car[]> => {
  const response = await fetch('/api/cars', {
    headers: {
      'Authorization': `Bearer ${authStore.token}`,
    },
  });
  if (!response.ok) {
    throw new Error('Network response was not ok');
  }
  const json = await response.json();
  return json.data;
};

const { isPending, isError, data, error } = useQuery({
  queryKey: ['cars'],
  queryFn: fetchCars,
});

const handleAddSuccess = () => {
  queryClient.invalidateQueries({ queryKey: ['cars'] });
};
</script>

<template>
  <div class="min-h-screen bg-gray-50 p-4">
    <header class="mb-6 flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Fleet Management</h1>
        <p class="text-gray-600">Manage your vehicles and inspections</p>
      </div>
      <button @click="isAddModalOpen = true" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors">
        Add Vehicle
      </button>
    </header>

    <!-- Loading State -->
    <div v-if="isPending" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="n in 3" :key="n" class="bg-white rounded-lg shadow-md h-80 animate-pulse">
        <div class="h-48 bg-gray-200 rounded-t-lg"></div>
        <div class="p-4 space-y-3">
          <div class="h-6 bg-gray-200 rounded w-3/4"></div>
          <div class="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="isError" class="p-4 bg-red-50 text-red-700 rounded-lg">
      Error loading fleet: {{ error?.message }}
    </div>

    <!-- Empty State -->
    <div v-else-if="!data || data.length === 0" class="text-center py-12 bg-white rounded-lg shadow-sm">
      <div class="text-gray-400 mb-4">
        <svg class="w-16 h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900 mb-2">No vehicles found</h3>
      <p class="text-gray-500 mb-6">Get started by adding your first vehicle to the fleet.</p>
      <button @click="isAddModalOpen = true" class="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors">
        Add your first Car
      </button>
    </div>

    <!-- Data State -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <CarCard
        v-for="car in data"
        :key="car.id"
        :car="car"
      />
    </div>

    <!-- Add Car Modal -->
    <AddCarModal
      :is-open="isAddModalOpen"
      @close="isAddModalOpen = false"
      @success="handleAddSuccess"
    />
  </div>
</template>
