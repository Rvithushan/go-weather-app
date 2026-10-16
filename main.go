package main


import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

//WeatherResponse maps the exact Json structure sent by the Open-Metep API.
type WeatherResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		WindSpeed float64 `json:"windspeed"`
	} `json:"current_weather"`
}


// GeocodingResponse maps the JSON from Open-Meteo Geocoding API
type GeocodingResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
	} `json:"results"`
}



// PageData holds the processed data we will eventually pass to our HTML page.
type PageData struct {
	City string
	Temperature float64
	WindSpeed float64
	Error string
}


// fetchCoordinates calls Open-Meteo Geocoding API to convert a city name to coordinates
func fetchCoordinates(city string) (float64, float64, string, error) {
	// 1. Construct the Geocoding API URL with the searched city name
	apiURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json", city)

	// 2. Perform the HTTP GET request
	resp, err := http.Get(apiURL)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to connect to geocoding API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, "", fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}

	// 3. Decode the JSON response into our GeocodingResponse struct
	var geoRes GeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoRes); err != nil {
		return 0, 0, "", fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	// 4. Check if any matching city results were found
	if len(geoRes.Results) == 0 {
		return 0, 0, "", fmt.Errorf("no location found for '%s'", city)
	}

	// Extract details for the first matching result
	location := geoRes.Results[0]
	displayName := fmt.Sprintf("%s, %s", location.Name, location.Country)

	return location.Latitude, location.Longitude, displayName, nil
}


// fetchWeather makes an HTTP GET request to the Open-Meteo API
func fetchWeather(lat, lon string) (*WeatherResponse, error) {
	// 1. Build the API endpoint URL with parameters
	apiURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current_weather=true", lat, lon)

	// 2. Perform the outbound HTTP GET request
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to reach weather API: %w", err)
	}
	// Always close the response body when done to prevent memory/connection leaks
	defer resp.Body.Close()

	// 3. Ensure the API responded with a 200 OK status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status code %d", resp.StatusCode)
	}

	// 4. Decode the raw JSON response directly into our WeatherResponse struct
	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &weather, nil
}


// weatherHandler processes HTTP requests for the web server
func weatherHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Unable to load HTML template", http.StatusInternalServerError)
		return
	}

	// 1. Read the 'city' query parameter from the URL (?city=Name)
	cityParam := r.URL.Query().Get("city")

	// Default to Colombo if no city parameter is provided
	if cityParam == "" {
		cityParam = "Colombo"
	}

	data := PageData{}

	// 2. Fetch coordinates for the requested city
	lat, lon, displayName, err := fetchCoordinates(cityParam)
	if err != nil {
		data.Error = err.Error()
		tmpl.Execute(w, data)
		return
	}

	data.City = displayName

	// 3. Fetch weather using the retrieved coordinates (convert float64 to string)
	latStr := fmt.Sprintf("%f", lat)
	lonStr := fmt.Sprintf("%f", lon)

	weather, err := fetchWeather(latStr, lonStr)
	if err != nil {
		data.Error = err.Error()
	} else {
		data.Temperature = weather.CurrentWeather.Temperature
		data.WindSpeed = weather.CurrentWeather.WindSpeed
	}

	// 4. Render index.html with the populated PageData
	tmpl.Execute(w, data)
}

// 4. BOTTOM: Entry point that starts the web server
func main() {
	http.HandleFunc("/", weatherHandler)
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
