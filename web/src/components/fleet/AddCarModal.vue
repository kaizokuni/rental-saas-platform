<script setup lang="ts">
import { ref } from 'vue';
import { useAuthStore } from '../../stores/auth';

const props = defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'success'): void;
}>();

const authStore = useAuthStore();
const isLoading = ref(false);
const error = ref<string | null>(null);

const form = ref({
  make: '',
  model: '',
  year: new Date().getFullYear(),
  license_plate: '',
  price_per_day: 0,
  status: 'available'
});

const imageFile = ref<File | null>(null);

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files[0]) {
    imageFile.value = target.files[0];
  }
};

const handleSubmit = async () => {
  if (!imageFile.value) {
    error.value = 'Please select an image';
    return;
  }

  isLoading.value = true;
  error.value = null;

  try {
    const formData = new FormData();
    formData.append('make', form.value.make);
    formData.append('model', form.value.model);
    formData.append('year', form.value.year.toString());
    formData.append('license_plate', form.value.license_plate);
    formData.append('price_per_day', form.value.price_per_day.toString());
    formData.append('status', form.value.status);
    formData.append('image', imageFile.value);

    const response = await fetch('/api/cars', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${authStore.token}`,
      },
      body: formData,
    });

    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.error || 'Failed to create car');
    }

    // Reset form
    form.value = {
      make: '',
      model: '',
      year: new Date().getFullYear(),
      license_plate: '',
      price_per_day: 0,
      status: 'available'
    };
    imageFile.value = null;
    
    emit('success');
    emit('close');
  } catch (e: any) {
    error.value = e.message;
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
    <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
      <!-- Background overlay -->
      <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" aria-hidden="true" @click="$emit('close')"></div>

      <!-- Modal panel -->
      <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
      <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
        <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
          <div class="sm:flex sm:items-start">
            <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left w-full">
              <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                Add New Vehicle
              </h3>
              <div class="mt-4">
                <form @submit.prevent="handleSubmit" class="space-y-4">
                  <!-- Error Message -->
                  <div v-if="error" class="bg-red-50 text-red-700 p-3 rounded-md text-sm">
                    {{ error }}
                  </div>

                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-gray-700">Make</label>
                      <input v-model="form.make" type="text" required class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm border p-2">
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700">Model</label>
                      <input v-model="form.model" type="text" required class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm border p-2">
                    </div>
                  </div>

                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-gray-700">Year</label>
                      <input v-model.number="form.year" type="number" required class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm border p-2">
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700">License Plate</label>
                      <input v-model="form.license_plate" type="text" required class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm border p-2">
                    </div>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-gray-700">Price Per Day (Cents)</label>
                    <div class="mt-1 relative rounded-md shadow-sm">
                      <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <span class="text-gray-500 sm:text-sm">$</span>
                      </div>
                      <input v-model.number="form.price_per_day" type="number" required min="0" class="focus:ring-blue-500 focus:border-blue-500 block w-full pl-7 pr-12 sm:text-sm border-gray-300 rounded-md border p-2" placeholder="0.00">
                    </div>
                    <p class="mt-1 text-xs text-gray-500">Enter amount in cents (e.g. 10000 = $100.00)</p>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-gray-700">Vehicle Image</label>
                    <input @change="handleFileChange" type="file" accept="image/*" required class="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100">
                  </div>

                  <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                    <button type="submit" :disabled="isLoading" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm disabled:opacity-50">
                      <span v-if="isLoading">Adding...</span>
                      <span v-else>Add Vehicle</span>
                    </button>
                    <button type="button" @click="$emit('close')" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:mt-0 sm:w-auto sm:text-sm">
                      Cancel
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
