package models

// Filters contiene parámetros de paginación y filtros
type Filters struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// NewFilters crea un objeto Filters con valores por defecto
func NewFilters(page, pageSize int) Filters {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // Límite máximo para evitar abuse
	}
	return Filters{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset calcula el offset para SQL
func (f Filters) Offset() int {
	return (f.Page - 1) * f.PageSize
}

// Metadata contiene información sobre la paginación
type Metadata struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	FirstPage    int `json:"first_page"`
	LastPage     int `json:"last_page"`
	TotalRecords int `json:"total_records"`
}

// CalculateMetadata calcula los metadatos de paginación
func CalculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{
			CurrentPage:  1,
			PageSize:     pageSize,
			FirstPage:    1,
			LastPage:     1,
			TotalRecords: 0,
		}
	}

	lastPage := (totalRecords + pageSize - 1) / pageSize
	if lastPage < 1 {
		lastPage = 1
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     lastPage,
		TotalRecords: totalRecords,
	}
}

// PaginatedResponse es la estructura genérica para respuestas paginadas
type PaginatedResponse struct {
	Items    interface{} `json:"items"`
	Metadata Metadata    `json:"metadata"`
}
