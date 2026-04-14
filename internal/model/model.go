package model

import "time"

type User struct {
	ID            int64      `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	Nickname      string     `json:"nickname"`
	AvatarURL     string     `json:"avatar_url"`
	Status        int16      `json:"status"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Category struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url"`
	CoverURL    string    `json:"cover_url"`
	SortOrder   int       `json:"sort_order"`
	IsHot       bool      `json:"is_hot"`
	Status      int16     `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Tea struct {
	ID             int64     `json:"id"`
	CategoryID     int64     `json:"category_id"`
	Name           string    `json:"name"`
	ShopName       string    `json:"shop_name"`
	Subtitle       string    `json:"subtitle"`
	Description    string    `json:"description"`
	ImageURL       string    `json:"image_url"`
	Tags           []string  `json:"tags"`
	PriceMinCent   int       `json:"price_min_cent"`
	PriceMaxCent   int       `json:"price_max_cent"`
	Currency       string    `json:"currency"`
	SugarLevel     string    `json:"sugar_level"`
	IsRecommended  bool      `json:"is_recommended"`
	Status         int16     `json:"status"`
	ViewCount      int64     `json:"view_count"`
	AvgRating      float64   `json:"avg_rating"`
	ReviewCount    int64     `json:"review_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type Review struct {
	ID        int64     `json:"id"`
	TeaID     int64     `json:"tea_id"`
	UserID    int64     `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	Stars     int16     `json:"stars"`
	Title     string    `json:"title"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type LeaderboardRow struct {
	Rank        int     `json:"rank"`
	TeaID       int64   `json:"tea_id"`
	Name        string  `json:"name"`
	ShopName    string  `json:"shop_name"`
	AvgRating   float64 `json:"avg_rating"`
	ReviewCount int64   `json:"review_count"`
}
