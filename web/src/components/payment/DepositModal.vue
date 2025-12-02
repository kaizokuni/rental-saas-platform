<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { loadStripe } from '@stripe/stripe-js';
import client from '../../api/client';

const props = defineProps<{
  bookingId: string;
  amountCents: number;
  isOpen: boolean;
}>();

const emit = defineEmits(['close', 'success']);

const stripe = ref<any>(null);
const elements = ref<any>(null);
const cardElement = ref<any>(null);
const isLoading = ref(false);
const error = ref('');

onMounted(async () => {
  stripe.value = await loadStripe(import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY);
  elements.value = stripe.value.elements();
  cardElement.value = elements.value.create('card');
  cardElement.value.mount('#card-element');
});

const handlePayment = async () => {
  isLoading.value = true;
  error.value = '';

  try {
    // 1. Get Client Secret from Backend
    const { data } = await client.post('/api/payments/intent', {
      booking_id: props.bookingId,
      amount: props.amountCents,
    });

    // 2. Confirm Card Payment (Manual Capture)
    const result = await stripe.value.confirmCardPayment(data.client_secret, {
      payment_method: {
        card: cardElement.value,
      },
    });

    if (result.error) {
      error.value = result.error.message;
    } else {
      if (result.paymentIntent.status === 'requires_capture') {
        emit('success');
        emit('close');
      } else {
        error.value = 'Payment status unexpected: ' + result.paymentIntent.status;
      }
    }
  } catch (err: any) {
    error.value = err.response?.data || err.message || 'Payment failed';
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
    <div class="bg-white rounded-lg shadow-xl w-full max-w-md p-6">
      <h2 class="text-xl font-bold mb-4">Authorize Deposit</h2>
      <p class="text-gray-600 mb-6">
        A hold of ${{ (amountCents / 100).toFixed(2) }} will be placed on your card.
      </p>

      <div class="mb-6">
        <div id="card-element" class="p-3 border border-gray-300 rounded-md"></div>
      </div>

      <div v-if="error" class="mb-4 text-red-600 text-sm">
        {{ error }}
      </div>

      <div class="flex justify-end space-x-3">
        <button
          @click="$emit('close')"
          class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-md"
          :disabled="isLoading"
        >
          Cancel
        </button>
        <button
          @click="handlePayment"
          class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          :disabled="isLoading"
        >
          <span v-if="isLoading">Processing...</span>
          <span v-else>Authorize Funds</span>
        </button>
      </div>
    </div>
  </div>
</template>
