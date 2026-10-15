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

// PageData holds the processed data we will eventually pass to our HTML page.
type PageData struct {
	City string
	Temperature float64
	WindSpeed float64
	Error string
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


// 3. MIDDLE/LOWER: The HTTP Handler (THIS IS WHERE YOU LOAD index.html)
func weatherHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the external HTML file
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Unable to load HTML template", http.StatusInternalServerError)
		return
	}

	data := PageData{City: "Colombo"}

	// Fetch weather data for Colombo coordinates
	weather, err := fetchWeather("6.9271", "79.8612")
	if err != nil {
		data.Error = err.Error()
	} else {
		data.Temperature = weather.CurrentWeather.Temperature
		data.WindSpeed = weather.CurrentWeather.WindSpeed
	}

	// Render the template with data and send it to the browser
	tmpl.Execute(w, data)
}

// 4. BOTTOM: Entry point that starts the web server
func main() {
	http.HandleFunc("/", weatherHandler)
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
