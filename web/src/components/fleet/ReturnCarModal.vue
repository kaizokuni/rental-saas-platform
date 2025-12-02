<script setup lang="ts">
import { ref } from 'vue';
import client from '../../api/client';

const props = defineProps<{
  bookingId: string;
  isOpen: boolean;
}>();

const emit = defineEmits(['close', 'success']);

const finalOdometer = ref<number | null>(null);
const damageCost = ref<number | null>(null);
const isLoading = ref(false);
const error = ref('');

const handleReturn = async () => {
  if (finalOdometer.value === null) {
    error.value = 'Please enter final odometer reading';
    return;
  }

  isLoading.value = true;
  error.value = '';

  try {
    await client.post(`/api/bookings/${props.bookingId}/return`, {
      final_odometer: finalOdometer.value,
      damage_cost_cents: (damageCost.value || 0) * 100, // Convert to cents
    });
    emit('success');
    emit('close');
  } catch (err: any) {
    error.value = err.response?.data || err.message || 'Return failed';
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
    <div class="bg-white rounded-lg shadow-xl w-full max-w-md p-6">
      <h2 class="text-xl font-bold mb-4">Return Car</h2>
      
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700">Final Odometer</label>
          <input
            type="number"
            v-model="finalOdometer"
            class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="e.g. 12050"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700">Damage Assessment ($)</label>
          <input
            type="number"
            v-model="damageCost"
            class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="0.00"
          />
        </div>

        <div v-if="error" class="text-red-600 text-sm">
          {{ error }}
        </div>

        <div class="flex justify-end space-x-3 mt-6">
          <button
            @click="$emit('close')"
            class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-md"
            :disabled="isLoading"
          >
            Cancel
          </button>
          <button
            @click="handleReturn"
            class="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50"
            :disabled="isLoading"
          >
            <span v-if="isLoading">Processing...</span>
            <span v-else>Complete Return</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
