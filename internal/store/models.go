package store

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    string
}

type Restaurant struct {
	ID                int64
	OwnerUserID       int64
	Name              string
	Slug              string
	Currency          string
	TargetFoodCostPct float64
	CreatedAt         string
}

type Menu struct {
	ID           int64
	RestaurantID int64
	Name         string
	Active       bool
	SortOrder    int
}

type Section struct {
	ID        int64
	MenuID    int64
	Name      string
	SortOrder int
}

type Item struct {
	ID            int64
	SectionID     int64
	RestaurantID  int64
	Name          string
	Description   string
	PriceCents    int64
	Active        bool
	SortOrder     int
	UpdatedAt     string
}

type Ingredient struct {
	ID            int64
	RestaurantID  int64
	Name          string
	Unit          string
	UnitCostCents float64 // cents per unit; fractional (e.g. 1.2 == $0.012)
	Supplier      string
	YieldPct      float64
	UpdatedAt     string
}

type RecipeLine struct {
	ID           int64
	ItemID       int64
	IngredientID int64
	Quantity     float64

	// joined fields, populated by some queries
	IngredientName string
	Unit           string
	UnitCostCents  float64
	YieldPct       float64
}

type SalesLine struct {
	ID           int64
	RestaurantID int64
	ImportID     *int64
	ItemID       *int64
	ItemName     string
	Quantity     float64
	SaleDate     string // YYYY-MM-DD
}
