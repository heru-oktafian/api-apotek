package models

// Menu represents the structure of each menu item in the JSON file.
type Menu struct {
	UserRole string `json:"user_role"`
	Details  []struct {
		GroupMenu string      `json:"group_menu"`
		Title     string      `json:"title"`
		URL       string      `json:"url"`
		Method    string      `json:"method,omitempty"`
		Access    interface{} `json:"access"` // Can be string or []string
	} `json:"details"`
}

// MenuResponse represents the overall structure of your menus.json file
type MenuResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    []Menu `json:"data"`
}

// Struktur untuk menerima input dari body request
type RequestBody struct {
	Page   int    `json:"page"`
	Search string `json:"search"`
	Month  string `json:"month"`
}
