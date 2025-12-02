<script setup lang="ts">
import { ref, onMounted } from 'vue';
import client from '../api/client';
import StatsCard from '../components/dashboard/StatsCard.vue';
import DepositModal from '../components/payment/DepositModal.vue';
import ReturnCarModal from '../components/fleet/ReturnCarModal.vue';

// --- Types ---
interface RecentBooking {
  id: string;
  car_make: string;
  car_model: string;
  status: string;
  created_at: string;
}

interface DashboardStats {
  revenue_cents: number;
  utilization_pct: number;
  active_rentals: number;
  recent_bookings: RecentBooking[];
}

// --- State ---
const stats = ref<DashboardStats | null>(null);
const isLoading = ref(true);
const error = ref('');

// Modal State
const isDepositModalOpen = ref(false);
const isReturnModalOpen = ref(false);
const selectedBookingId = ref('');
const depositAmount = ref(50000); // Default $500.00 (in cents)

// --- Helpers ---

// Format cents (integer) to USD string (e.g. 15000 -> "$150.00")
const formatCurrency = (cents: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(cents / 100);
};

// Format ISO date string to readable date
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

// Returns Tailwind classes based on status
const getStatusClass = (status: string) => {
  const classes: Record<string, string> = {
    pending: 'text-yellow-700 bg-yellow-50 ring-yellow-600/20',
    active: 'text-green-700 bg-green-50 ring-green-600/20',
    confirmed: 'text-green-700 bg-green-50 ring-green-600/20',
    completed: 'text-gray-600 bg-gray-50 ring-gray-500/10',
    cancelled: 'text-red-700 bg-red-50 ring-red-600/10',
  };
  // Base badge styles + specific status style
  return `inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${classes[status] || 'text-gray-500 bg-gray-50 ring-gray-500/10'}`;
};

// --- Actions ---

const fetchStats = async () => {
  try {
    const { data } = await client.get('/api/dashboard/stats');
    stats.value = data;
  } catch (err: any) {
    error.value = 'Failed to load dashboard data';
    console.error(err);
  } finally {
    isLoading.value = false;
  }
};

const openDepositModal = (bookingId: string) => {
  selectedBookingId.value = bookingId;
  isDepositModalOpen.value = true;
};

const openReturnModal = (bookingId: string) => {
  selectedBookingId.value = bookingId;
  isReturnModalOpen.value = true;
</script>

<template>
  <div class="space-y-6">
    <div class="md:flex md:items-center md:justify-between">
      <div class="min-w-0 flex-1">
        <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight">
          Dashboard
        </h2>
      </div>
    </div>

    <div v-if="isLoading" class="flex justify-center py-20">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
    </div>

    <div v-else-if="error" class="rounded-md bg-red-50 p-4">
      <div class="flex">
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">Error loading dashboard</h3>
          <div class="mt-2 text-sm text-red-700">
            <p>{{ error }}</p>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="space-y-8">
      
      <div class="grid grid-cols-1 gap-5 sm:grid-cols-3">
        <StatsCard
          title="Revenue (This Month)"
          :value="formatCurrency(stats?.revenue_cents || 0)"
          icon="$"
        />
        <StatsCard
          title="Utilization Rate"
          :value="(stats?.utilization_pct || 0).toFixed(1) + '%'"
          icon="%"
        />
        <StatsCard
          title="Active Rentals"
          :value="stats?.active_rentals || 0"
          icon="#"
        />
      </div>

      <div class="overflow-hidden bg-white shadow sm:rounded-md">
        <div class="px-4 py-5 sm:px-6">
          <h3 class="text-base font-semibold leading-6 text-gray-900">Recent Activity</h3>
        </div>
        <ul role="list" class="divide-y divide-gray-200">
          <li v-for="booking in stats?.recent_bookings" :key="booking.id" class="px-4 py-4 sm:px-6 hover:bg-gray-50">
            <div class="flex items-center justify-between">
              
              <div class="flex flex-col truncate">
                <p class="truncate text-sm font-medium text-indigo-600">
                  {{ booking.car_make }} {{ booking.car_model }}
                </p>
                <div class="flex items-center mt-1">
                  <span class="truncate text-sm text-gray-500 mr-2">
                    {{ formatDate(booking.created_at) }}
                  </span>
                  <span :class="getStatusClass(booking.status)">
                    {{ booking.status }}
                  </span>
                </div>
              </div>

              <div class="flex space-x-2">
                <button
                  v-if="booking.status === 'pending'"
                  @click="openDepositModal(booking.id)"
                  class="rounded bg-indigo-50 px-2.5 py-1.5 text-xs font-semibold text-indigo-600 shadow-sm hover:bg-indigo-100"
                >
                  Authorize Deposit
                </button>

                <button
                  v-if="booking.status === 'active' || booking.status === 'confirmed'"
                  @click="openReturnModal(booking.id)"
                  class="rounded bg-green-50 px-2.5 py-1.5 text-xs font-semibold text-green-600 shadow-sm hover:bg-green-100"
                >
                  Return Car
                </button>

                <button
                  v-if="booking.status === 'completed'"
                  @click="downloadInvoice(booking.id)"
                  class="rounded bg-gray-50 px-2.5 py-1.5 text-xs font-semibold text-gray-600 shadow-sm hover:bg-gray-100"
                >
                  Invoice
                </button>
              </div>

            </div>
          </li>

          <li v-if="!stats?.recent_bookings?.length" class="px-4 py-10 text-center text-sm text-gray-500">
            No recent bookings found. Go to the Fleet page to book a car.
          </li>
        </ul>
      </div>
    </div>

    <DepositModal
      :is-open="isDepositModalOpen"
      :booking-id="selectedBookingId"
      :amount-cents="depositAmount"
      @close="isDepositModalOpen = false"
      @success="handleActionSuccess"
    />

    <ReturnCarModal
      :is-open="isReturnModalOpen"
      :booking-id="selectedBookingId"
      @close="isReturnModalOpen = false"
      @success="handleActionSuccess"
    />
  </div>
</template>