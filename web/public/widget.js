(function () {
    // 1. Find the container
    const container = document.getElementById('rental-widget');
    if (!container) {
        console.error('Rental Widget: Container #rental-widget not found.');
        return;
    }

    const tenantID = container.getAttribute('data-tenant-id');
    if (!tenantID) {
        console.error('Rental Widget: data-tenant-id attribute is missing.');
        container.innerHTML = '<p style="color:red">Error: Missing Tenant ID</p>';
        return;
    }

    const API_BASE = 'http://localhost:8080/api/public'; // Should be configurable or detected

    // 2. Inject CSS
    const style = document.createElement('style');
    style.textContent = `
        #rental-widget { font-family: sans-serif; max-width: 800px; margin: 0 auto; }
        .rw-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 20px; }
        .rw-card { border: 1px solid #ddd; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        .rw-image { width: 100%; height: 150px; object-fit: cover; background: #f0f0f0; }
        .rw-content { padding: 15px; }
        .rw-title { margin: 0 0 10px; font-size: 18px; font-weight: bold; }
        .rw-price { color: #2c7a7b; font-weight: bold; font-size: 16px; margin-bottom: 10px; }
        .rw-btn { background: #3182ce; color: white; border: none; padding: 10px 15px; border-radius: 4px; cursor: pointer; width: 100%; font-size: 14px; }
        .rw-btn:hover { background: #2c5282; }
        
        /* Modal */
        .rw-modal-overlay { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); display: flex; justify-content: center; align-items: center; z-index: 10000; }
        .rw-modal { background: white; padding: 20px; border-radius: 8px; width: 90%; max-width: 400px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        .rw-form-group { margin-bottom: 15px; }
        .rw-label { display: block; margin-bottom: 5px; font-weight: 500; font-size: 14px; }
        .rw-input { width: 100%; padding: 8px; border: 1px solid #ccc; border-radius: 4px; box-sizing: border-box; }
        .rw-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 20px; }
        .rw-btn-secondary { background: #e2e8f0; color: #4a5568; }
    `;
    document.head.appendChild(style);

    // 3. Fetch Cars
    async function fetchCars() {
        container.innerHTML = '<p>Loading cars...</p>';
        try {
            const response = await fetch(`${API_BASE}/cars?tenant_id=${tenantID}`);
            if (!response.ok) throw new Error('Failed to load cars');
            const result = await response.json();
            renderCars(result.data || []);
        } catch (err) {
            console.error(err);
            container.innerHTML = '<p style="color:red">Failed to load available cars.</p>';
        }
    }

    // 4. Render Cars
    function renderCars(cars) {
        if (cars.length === 0) {
            container.innerHTML = '<p>No cars available at the moment.</p>';
            return;
        }

        const grid = document.createElement('div');
        grid.className = 'rw-grid';

        cars.forEach(car => {
            const card = document.createElement('div');
            card.className = 'rw-card';

            const price = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(car.price_cents / 100);
            const image = car.image_url || 'https://via.placeholder.com/300x200?text=No+Image';

            card.innerHTML = `
                <img src="${image}" alt="${car.make} ${car.model}" class="rw-image">
                <div class="rw-content">
                    <h3 class="rw-title">${car.make} ${car.model} (${car.year})</h3>
                    <div class="rw-price">${price} / day</div>
                    <button class="rw-btn" data-id="${car.id}">Book Now</button>
                </div>
            `;

            card.querySelector('button').addEventListener('click', () => openBookingModal(car));
            grid.appendChild(card);
        });

        container.innerHTML = '';
        container.appendChild(grid);
    }

    // 5. Booking Modal
    function openBookingModal(car) {
        const overlay = document.createElement('div');
        overlay.className = 'rw-modal-overlay';

        const today = new Date().toISOString().split('T')[0];
        const tomorrow = new Date(Date.now() + 86400000).toISOString().split('T')[0];

        overlay.innerHTML = `
            <div class="rw-modal">
                <h3 class="rw-title">Book ${car.make} ${car.model}</h3>
                <form id="rw-booking-form">
                    <div class="rw-form-group">
                        <label class="rw-label">Full Name</label>
                        <input type="text" name="name" class="rw-input" required>
                    </div>
                    <div class="rw-form-group">
                        <label class="rw-label">Email</label>
                        <input type="email" name="email" class="rw-input" required>
                    </div>
                    <div class="rw-form-group">
                        <label class="rw-label">Start Date</label>
                        <input type="date" name="start_date" class="rw-input" min="${today}" required>
                    </div>
                    <div class="rw-form-group">
                        <label class="rw-label">End Date</label>
                        <input type="date" name="end_date" class="rw-input" min="${tomorrow}" required>
                    </div>
                    <div class="rw-actions">
                        <button type="button" class="rw-btn rw-btn-secondary" id="rw-cancel">Cancel</button>
                        <button type="submit" class="rw-btn">Submit Request</button>
                    </div>
                </form>
            </div>
        `;

        document.body.appendChild(overlay);

        const close = () => document.body.removeChild(overlay);
        document.getElementById('rw-cancel').addEventListener('click', close);

        document.getElementById('rw-booking-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = {
                car_id: car.id,
                customer_name: formData.get('name'),
                customer_email: formData.get('email'),
                start_date: new Date(formData.get('start_date')).toISOString(),
                end_date: new Date(formData.get('end_date')).toISOString()
            };

            try {
                const btn = e.target.querySelector('button[type="submit"]');
                btn.textContent = 'Submitting...';
                btn.disabled = true;

                const res = await fetch(`${API_BASE}/book?tenant_id=${tenantID}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });

                if (!res.ok) throw new Error(await res.text());

                alert('Booking request received! We will contact you shortly.');
                close();
            } catch (err) {
                alert('Booking failed: ' + err.message);
                e.target.querySelector('button[type="submit"]').textContent = 'Submit Request';
                e.target.querySelector('button[type="submit"]').disabled = false;
            }
        });
    }

    // Initialize
    fetchCars();
})();
