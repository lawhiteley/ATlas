const pinsElement = document.getElementById('pins');
const pinsData = JSON.parse(pinsElement.dataset.pins);
const { Longitude: long = 13, Latitude: lat = 42 } = window.savedPin || {}; // TODO: maybe default to geoloc
const map = new maplibregl.Map({
container: 'map',
style: 'https://tiles.basemaps.cartocdn.com/gl/voyager-gl-style/style.json',
center: [long, lat],
zoom: 3.3
});

map.on('style.load', () => {
    map.setProjection({
        type: 'globe'
    });
});

const popup = new maplibregl.Popup({
    closeButton: false,
    closeOnClick: false,
    maxWidth: '300px'
});

map.on('click', (e) => {
    if (Alpine.store('savedPin').hasPin) return;
    const coords = e.lngLat;
    
    const popupContent = document.createElement('div');
    popupContent.className = 'shadow-xl bg-base-200/90 backdrop-blur-sm border-base-300 rounded-box';
    popupContent.innerHTML = `
        <div class="card-body">
            <div class="card-title flex justify-between items-center">
                <h3 class="text-lg font-bold">Place a Pin</h3>
                <button class="btn btn-sm btn-circle btn-ghost close-popup">
                    <i class="ri-close-line text-lg"></i>
                </button>
            </div>
            <p class="text-sm text-base-content/70 mb-2">Click the button to place a pin at these coordinates:</p>
            <div class="bg-base-200 p-3 rounded-lg mb-4">
                <code class="text-sm">${coords.lng.toFixed(6)}, ${coords.lat.toFixed(6)}</code>
            </div>
            <div class="card-actions justify-end">
                <button class="btn btn-primary btn-sm place-pin-btn">
                    <i class="ri-map-pin-line mr-2"></i>
                    Place Pin
                </button>
            </div>
        </div>
    `;
    
    popup.setDOMContent(popupContent);
    popup.setLngLat(coords).addTo(map);
    
    popupContent.querySelector('.close-popup').addEventListener('click', () => {
        popup.remove();
    });
    
    popupContent.querySelector('.place-pin-btn').addEventListener('click', () => {
        addPin(coords, "hello");
        popup.remove();
    });
});

async function addPin(coords, description) {
    try {
        const response = await fetch('/pin', {
            method: "POST",
            headers: { 'Content-Type': 'application/x-www-form-urlencoded'},
            body: new URLSearchParams({
                longitude: coords.lng,
                latitude: coords.lat,
                description: description
            })
        });

        const pin = await response.json();
        pinsData[pin.Did] = pin; // TODO: remove this?
        Alpine.store('savedPin').data = pin;

        createMarkerForPin(pin);

        return pin;
    } catch (error) {
        console.error(`Failed to add pin: ${error}`)
    }
}

function createMarkerForPin(pin) {
    const newElement = document.createElement('div');
    newElement.className = 'pin';
    newElement.innerHTML = `
        <div class="avatar">
          <div class="w-10 rounded-full ring-2 ring-primary ring-offset-2 ring-offset-base-100">
            <img src="${pin.Avatar}" alt="${pin.Name}" class="object-cover" />
          </div>
        </div>
    `

    const marker = new maplibregl.Marker({ element: newElement, anchor: 'center' })
        .setLngLat([pin.Longitude, pin.Latitude])
        .addTo(map);

    pinsData[pin.Did].marker = marker; // TODO: only actually need to store reference to own marker
    marker._pin = pin;

    return marker;
}

// Paint pins on map
Object.values(pinsData).map(createMarkerForPin);

document.addEventListener('alpine:init', () => {
    Alpine.store('savedPin', {
        data: window.savedPin || null,

        get hasPin() {
            return this.data !== null;
        },
        
        async remove() {
            await fetch('/pin', { method: 'DELETE' });
            pinsData[this.data.Did].marker.remove();
            this.data = null;
        },

        flyToUserPin() {
            if (this.data) {
                map.flyTo({
                    center: [this.data.Longitude, this.data.Latitude],
                    speed: 0.8,
                    zoom: 9,  
                });
            }
        }
    });
});