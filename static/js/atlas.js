const pinsElement = document.getElementById('pins');
const pinsData = JSON.parse(pinsElement.dataset.pins);
const { Longitude: long = 13, Latitude: lat = 42 } = window.savedPin || {};
const map = new maplibregl.Map({
    container: 'map',
    style: 'https://tiles.basemaps.cartocdn.com/gl/voyager-gl-style/style.json',
    center: [long, lat],
    zoom: 3.3,
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
    if (Alpine.store('savedPin').hasPin || !document.cookie.includes(`oauth-session=`)) return;
    const coords = e.lngLat;
    
    const popupContent = document.createElement('div');
    popupContent.className = 'card shadow-xl bg-base-200/90 backdrop-blur-sm border-base-300 rounded-box';
    popupContent.innerHTML = `
        <fieldset class="card-body p-3 m-2">
            <div class="card-title flex justify-between items-center">
            <legend class="fieldset-legend">Place your Pin</legend>
                <button class="btn btn-sm btn-circle btn-ghost close-popup">
                    <i class="ri-close-line text-lg"></i>
                </button>
            </div>
            <label class="label">Description</label>
            <textarea placeholder="What are you doing here?" class="textarea textarea-s pin-description"></textarea>

            <label class="label">Website</label>
            <input type="text" class="input input-s pin-website" placeholder="https://luke.whiteley.io" />
            <button class="btn btn-primary btn-sm place-pin-btn">
                <i class="ri-map-pin-line mr-2"></i> Place Pin
            </button>
        </fieldset>
    `;
    
    popup.setDOMContent(popupContent);
    popup.setLngLat(coords).addTo(map);
    
    popupContent.querySelector('.close-popup').addEventListener('click', () => {
        popup.remove();
    });
    
    popupContent.querySelector('.place-pin-btn').addEventListener('click', () => {
        const description = popupContent.querySelector('.pin-description').value;
        const website = popupContent.querySelector('.pin-website').value;

        addPin(coords, description, website);
        popup.remove();
    });
});

async function addPin(coords, description, website) {
    try {
        const response = await fetch('/pin', {
            method: "POST",
            headers: { 'Content-Type': 'application/x-www-form-urlencoded'},
            body: new URLSearchParams({
                longitude: coords.lng,
                latitude: coords.lat,
                description: description,
                website: website
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

let openPin = null;

function createMarkerForPin(pin) {
    const newElement = document.createElement('div');
    newElement.className = 'pin cursor-pointer';
    newElement.innerHTML = `
        <div class="avatar">
          <div class="w-10 rounded-full ring-2 ring-primary ring-offset-2 ring-offset-base-100">
            <img src="${pin.Avatar}" alt="${pin.Name}" class="object-cover" />
          </div>
        </div>
    `;

    const marker = new maplibregl.Marker({ element: newElement, anchor: 'center' })
        .setLngLat([pin.Longitude, pin.Latitude])
        .addTo(map);

    pinsData[pin.Did].marker = marker;
    marker._pin = pin;

    const popupContent = document.createElement('div');
    popupContent.className = 'p-0 overflow-hidden rounded-box w-80';
    popupContent.innerHTML = `
        <div class="card bg-base-100 shadow-xl">
            <button class="btn btn-circle btn-xs absolute right-3 top-3 z-10 close-popup hover:bg-primary">
                <i class="ri-close-line text-lg"></i>
            </button>
            <figure>
                <img src="${pin.Avatar}" alt="${pin.Name}" class="h-40 w-full object-cover" />
            </figure>
            <div class="card-body p-4">
                <div class="flex-1 min-w-0">
                    <div class="font-semibold text-base truncate">${pin.Name}</div>
                    <div class="text-sm text-base-content/60 truncate">${pin.Handle}</div>
                </div>
                ${pin.Description ? `<p>${pin.Description}</p>` : ''}
                ${pin.Website ? `<a class="link link-accent">${pin.Website}</a>` : ''}
            </div>
        </div>
    `;

    const popup = new maplibregl.Popup({
        offset: 25,
        closeButton: false,
        closeOnClick: true,
        className: 'maplibregl-popup-content rounded-box shadow-2xl p-0'
    }).setDOMContent(popupContent);

    newElement.addEventListener('click', (e) => {
        e.stopPropagation();

        if (openPin && openPin !== popup) {
            openPin.remove();
        }

        marker.setPopup(popup).togglePopup();

        openPin = popup;

        setTimeout(() => {
            const closeButton = popupContent.querySelector('.close-popup');
            if (closeButton) {
                closeButton.addEventListener('click', () => {
                    popup.remove();
                    openPin = null;
                });
            }
        }, 0);
    });

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