package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MercadoLibreService maneja las operaciones de OAuth2 con Mercado Libre
type MercadoLibreService struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// MLTokenResponse representa la respuesta de Mercado Libre al intercambiar el código
type MLTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // segundos
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	UserID       int64  `json:"user_id"`
}

// MLUserInfo representa información básica del usuario de Mercado Libre
type MLUserInfo struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

// NewMercadoLibreService crea una nueva instancia del servicio
func NewMercadoLibreService(clientID, clientSecret, redirectURI string) *MercadoLibreService {
	return &MercadoLibreService{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
	}
}

// GetAuthorizationURL genera la URL de autorización de Mercado Libre
// El parámetro state se puede usar para pasar el user_id de nuestro sistema
func (s *MercadoLibreService) GetAuthorizationURL(state string) string {
	baseURL := "https://auth.mercadolibre.com.ar/authorization"
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", s.ClientID)
	params.Add("redirect_uri", s.RedirectURI)
	params.Add("state", state)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// ExchangeCodeForToken intercambia el código de autorización por tokens
func (s *MercadoLibreService) ExchangeCodeForToken(code string) (*MLTokenResponse, error) {
	tokenURL := "https://api.mercadolibre.com/oauth/token"

	// Preparar el body de la petición
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", s.ClientID)
	data.Set("client_secret", s.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", s.RedirectURI)

	// Hacer la petición POST
	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error making token request: %w", err)
	}
	defer resp.Body.Close()

	// Leer el body de la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Verificar el código de estado
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mercadolibre API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parsear la respuesta JSON
	var tokenResp MLTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("error parsing token response: %w", err)
	}

	return &tokenResp, nil
}

// RefreshAccessToken refresca un access_token usando el refresh_token
func (s *MercadoLibreService) RefreshAccessToken(refreshToken string) (*MLTokenResponse, error) {
	tokenURL := "https://api.mercadolibre.com/oauth/token"

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", s.ClientID)
	data.Set("client_secret", s.ClientSecret)
	data.Set("refresh_token", refreshToken)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error refreshing token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mercadolibre API error (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp MLTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("error parsing token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo obtiene información del usuario de Mercado Libre usando el access_token
func (s *MercadoLibreService) GetUserInfo(accessToken string) (*MLUserInfo, error) {
	userURL := "https://api.mercadolibre.com/users/me"

	req, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making user info request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mercadolibre API error (status %d): %s", resp.StatusCode, string(body))
	}

	var userInfo MLUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("error parsing user info: %w", err)
	}

	return &userInfo, nil
}

// MLOrder representa una orden de Mercado Libre
type MLOrder struct {
	ID     int64  `json:"id"`
	Status string `json:"status"` // confirmed, payment_required, payment_in_process, partially_paid, paid, cancelled
	Buyer  struct {
		ID        int64  `json:"id"`
		Nickname  string `json:"nickname"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     struct {
			AreaCode string `json:"area_code"`
			Number   string `json:"number"`
		} `json:"phone"`
	} `json:"buyer"`
	OrderItems []struct {
		Item struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			CategoryID  string `json:"category_id"`
			VariationID int64  `json:"variation_id"`
			SellerSKU   string `json:"seller_custom_field"` // SKU del vendedor
		} `json:"item"`
		Quantity      int     `json:"quantity"`
		UnitPrice     float64 `json:"unit_price"`
		FullUnitPrice float64 `json:"full_unit_price"`
		SalePrice     float64 `json:"sale_fee"`
		ListingTypeID string  `json:"listing_type_id"`
	} `json:"order_items"`
	Payments []struct {
		ID     int64   `json:"id"`
		Status string  `json:"status"`
		Total  float64 `json:"transaction_amount"`
	} `json:"payments"`
	Shipping struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	} `json:"shipping"`
	DateCreated string  `json:"date_created"`
	DateClosed  string  `json:"date_closed"`
	TotalAmount float64 `json:"total_amount"`
	CurrencyID  string  `json:"currency_id"`
}

// GetOrder obtiene los detalles completos de una orden de Mercado Libre
func (s *MercadoLibreService) GetOrder(orderID int64, accessToken string) (*MLOrder, error) {
	orderURL := fmt.Sprintf("https://api.mercadolibre.com/orders/%d", orderID)

	req, err := http.NewRequest("GET", orderURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making order request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mercadolibre API error (status %d): %s", resp.StatusCode, string(body))
	}

	var order MLOrder
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("error parsing order: %w", err)
	}

	return &order, nil
}
