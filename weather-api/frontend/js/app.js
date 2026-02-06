// ==================== DOM Elements ====================
const weatherForm = document.getElementById('weatherForm');
const cityInput = document.getElementById('cityInput');
const searchBtn = document.querySelector('.search-btn');
const btnLoader = document.getElementById('btnLoader');
const loadingContainer = document.getElementById('loadingContainer');
const errorContainer = document.getElementById('errorContainer');
const errorTitle = document.getElementById('errorTitle');
const errorMessage = document.getElementById('errorMessage');
const weatherContainer = document.getElementById('weatherContainer');
const exampleBtns = document.querySelectorAll('.example-btn');

// Weather display elements
const cityNameEl = document.getElementById('cityName');
const countryEl = document.getElementById('country');
const lastUpdatedEl = document.getElementById('lastUpdated');
const cacheBadgeEl = document.getElementById('cacheBadge');
const weatherIconEl = document.getElementById('weatherIcon');
const tempValueEl = document.getElementById('tempValue');
const weatherConditionEl = document.getElementById('weatherCondition');
const feelsLikeEl = document.getElementById('feelsLike');
const humidityEl = document.getElementById('humidity');
const windSpeedEl = document.getElementById('windSpeed');
const pressureEl = document.getElementById('pressure');
const visibilityEl = document.getElementById('visibility');
const weatherDescriptionEl = document.getElementById('weatherDescription');

// ==================== Event Listeners ====================
weatherForm.addEventListener('submit', handleSearch);
exampleBtns.forEach(btn => {
    btn.addEventListener('click', () => {
        cityInput.value = btn.dataset.city;
        weatherForm.requestSubmit();
    });
});

// ==================== Main Functions ====================
async function handleSearch(e) {
    e.preventDefault();

    const city = cityInput.value.trim();

    if (!city) {
        showError('Please enter a city name');
        return;
    }

    // Show loading state
    setLoading(true);

    try {
        const weatherData = await fetchWeather(city);
        displayWeather(weatherData);
    } catch (error) {
        console.error('Error:', error);
        showError(error.message || 'Failed to fetch weather data');
    } finally {
        setLoading(false);
    }
}

async function fetchWeather(city) {
    const response = await fetch(`/api/weather?city=${encodeURIComponent(city)}`);

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }

    return await response.json();
}

function displayWeather(data) {
    // Hide other containers
    hideAllContainers();
    weatherContainer.style.display = 'block';

    // Parse city and country from resolvedAddress (e.g., "Jakarta, Indonesia")
    const addressParts = data.resolvedAddress?.split(',') || [data.address || 'Unknown'];
    cityNameEl.textContent = addressParts[0]?.trim() || 'Unknown';
    countryEl.textContent = addressParts[1]?.trim() || '';

    // Current conditions
    const current = data.currentConditions;

    // Temperature
    tempValueEl.textContent = Math.round(current.temp);

    // Weather icon based on conditions
    weatherIconEl.textContent = getWeatherIcon(current.conditions);

    // Weather condition
    weatherConditionEl.textContent = current.conditions || 'Unknown';

    // Feels like
    feelsLikeEl.textContent = `Feels like ${Math.round(current.feelslike)}°C`;

    // Details
    humidityEl.textContent = `${Math.round(current.humidity)}%`;
    windSpeedEl.textContent = `${Math.round(current.windspeed)} km/h`;
    pressureEl.textContent = `${Math.round(current.pressure)} hPa`;
    visibilityEl.textContent = `${Math.round(current.visibility)} km`;

    // Description
    weatherDescriptionEl.textContent = current.description || 'No description available';

    // Last updated
    const lastUpdatedTime = new Date(current.datetime);
    lastUpdatedEl.textContent = `Last updated: ${formatDateTime(lastUpdatedTime)}`;

    // Cache badge
    cacheBadgeEl.style.display = data.cached ? 'inline-block' : 'none';
}

function setLoading(isLoading) {
    if (isLoading) {
        searchBtn.classList.add('btn-loading');
        searchBtn.disabled = true;
        loadingContainer.style.display = 'block';
        hideAllContainers();
    } else {
        searchBtn.classList.remove('btn-loading');
        searchBtn.disabled = false;
        loadingContainer.style.display = 'none';
    }
}

function showError(title, message = '') {
    hideAllContainers();
    errorContainer.style.display = 'block';
    errorTitle.textContent = title;
    errorMessage.textContent = message;
}

function hideAllContainers() {
    loadingContainer.style.display = 'none';
    errorContainer.style.display = 'none';
    weatherContainer.style.display = 'none';
}

// ==================== Helper Functions ====================
function getWeatherIcon(condition) {
    const conditionLower = condition.toLowerCase();

    if (conditionLower.includes('clear') || conditionLower.includes('sunny')) return '☀️';
    if (conditionLower.includes('cloudy') || conditionLower.includes('cloud')) return '☁️';
    if (conditionLower.includes('rain') || conditionLower.includes('drizzle')) return '🌧️';
    if (conditionLower.includes('thunderstorm') || conditionLower.includes('thunder')) return '⛈️';
    if (conditionLower.includes('snow')) return '❄️';
    if (conditionLower.includes('fog') || conditionLower.includes('mist')) return '🌫️';
    if (conditionLower.includes('wind')) return '💨';

    return '🌤️';
}

function formatDateTime(date) {
    return date.toLocaleString('en-US', {
        weekday: 'short',
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

// ==================== Initialize ====================
// Focus on input when page loads
window.addEventListener('load', () => {
    cityInput.focus();

    // Check if there's a city in URL params (for direct links)
    const urlParams = new URLSearchParams(window.location.search);
    const cityParam = urlParams.get('city');
    if (cityParam) {
        cityInput.value = cityParam;
        weatherForm.requestSubmit();
    }
});