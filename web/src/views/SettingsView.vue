<script setup lang="ts">
import { ref, onMounted } from 'vue';
import client from '../api/client';

interface Webhook {
  id: string;
  url: string;
  events: string[];
  active: boolean;
  secret_key: string;
}

const webhooks = ref<Webhook[]>([]);
const newWebhookUrl = ref('');
const selectedEvents = ref<string[]>([]);
const availableEvents = ['booking.created', 'booking.completed', 'payment.captured', 'car.status_changed'];
const isLoading = ref(false);
const error = ref('');

const fetchWebhooks = async () => {
  try {
    const { data } = await client.get('/api/settings/webhooks');
    webhooks.value = data || [];
  } catch (err) {
    console.error('Failed to fetch webhooks', err);
  }
};

const addWebhook = async () => {
  if (!newWebhookUrl.value || selectedEvents.value.length === 0) {
    error.value = 'URL and at least one event are required';
    return;
  }

  isLoading.value = true;
  error.value = '';

  try {
    const { data } = await client.post('/api/settings/webhooks', {
      url: newWebhookUrl.value,
      events: selectedEvents.value,
    });
    webhooks.value.push(data);
    newWebhookUrl.value = '';
    selectedEvents.value = [];
  } catch (err: any) {
    error.value = err.response?.data || 'Failed to add webhook';
  } finally {
    isLoading.value = false;
  }
};

onMounted(() => {
  fetchWebhooks();
});
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-gray-900">Settings</h1>

    <!-- Webhooks Section -->
    <div class="bg-white shadow sm:rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900">Webhooks</h3>
        <div class="mt-2 max-w-xl text-sm text-gray-500">
          <p>Register webhooks to receive real-time event notifications.</p>
        </div>
        
        <div class="mt-5 space-y-4">
          <!-- Add Webhook Form -->
          <div class="space-y-3">
            <div>
              <label class="block text-sm font-medium text-gray-700">Endpoint URL</label>
              <input
                type="url"
                v-model="newWebhookUrl"
                class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                placeholder="https://api.yourapp.com/webhooks"
              />
            </div>
            
            <div>
              <label class="block text-sm font-medium text-gray-700">Events</label>
              <div class="mt-2 space-y-2">
                <div v-for="event in availableEvents" :key="event" class="flex items-center">
                  <input
                    :id="event"
                    type="checkbox"
                    :value="event"
                    v-model="selectedEvents"
                    class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                  <label :for="event" class="ml-2 block text-sm text-gray-900">
                    {{ event }}
                  </label>
                </div>
              </div>
            </div>

            <div v-if="error" class="text-red-600 text-sm">
              {{ error }}
            </div>

            <button
              @click="addWebhook"
              class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              :disabled="isLoading"
            >
              Add Webhook
            </button>
          </div>

          <!-- Webhook List -->
          <div class="mt-6 border-t border-gray-200 pt-4">
            <h4 class="text-sm font-medium text-gray-900 mb-4">Registered Webhooks</h4>
            <ul role="list" class="divide-y divide-gray-200">
              <li v-for="webhook in webhooks" :key="webhook.id" class="py-4">
                <div class="flex items-center justify-between">
                  <div class="text-sm">
                    <p class="font-medium text-gray-900">{{ webhook.url }}</p>
                    <p class="text-gray-500">{{ webhook.events.join(', ') }}</p>
                    <p class="text-xs text-gray-400 mt-1">Secret: {{ webhook.secret_key }}</p>
                  </div>
                  <span
                    class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full"
                    :class="webhook.active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
                  >
                    {{ webhook.active ? 'Active' : 'Inactive' }}
                  </span>
                </div>
              </li>
              <li v-if="webhooks.length === 0" class="text-sm text-gray-500 italic">
                No webhooks registered.
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
