package purchase_recommender

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Piyushhbhutoria/tote-assignment/internal/models"
	"github.com/Piyushhbhutoria/tote-assignment/pkg/database"
)

// Plugin implements the purchase recommender plugin
type Plugin struct {
	db     *database.Connection
	active bool
	config map[string]interface{}
}

// New creates a new purchase recommender plugin
func New(db *database.Connection) *Plugin {
	return &Plugin{
		db:     db,
		active: false,
		config: make(map[string]interface{}),
	}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "purchase_recommender"
}

// Description returns the plugin description
func (p *Plugin) Description() string {
	return "Analyzes basket items and provides purchase recommendations"
}

// IsActive returns whether the plugin is active
func (p *Plugin) IsActive() bool {
	return p.active
}

// SetActive enables or disables the plugin
func (p *Plugin) SetActive(active bool) {
	p.active = active
}

// Configure applies the plugin configuration
func (p *Plugin) Configure(config map[string]interface{}) error {
	p.config = config
	return nil
}

// ProcessEvent handles item addition events
func (p *Plugin) ProcessEvent(ctx context.Context, event *models.Event) ([]*models.Event, error) {
	if !p.active {
		return nil, nil
	}

	if event.Type != models.EventAddItem {
		return nil, nil
	}

	return p.handleItemAdded(ctx, event)
}

func (p *Plugin) handleItemAdded(ctx context.Context, event *models.Event) ([]*models.Event, error) {
	var payload struct {
		BasketID   string  `json:"basket_id"`
		ItemID     string  `json:"item_id"`
		TerminalID string  `json:"terminal_id"`
		StoreID    string  `json:"store_id"`
		Price      float64 `json:"price"`
	}

	data, err := json.Marshal(event.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	// Get recommendations for the added item
	recommendations, err := p.getRecommendations(ctx, payload.ItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %v", err)
	}

	if len(recommendations) == 0 {
		return nil, nil
	}

	// Create recommendation event
	recommendEvent := &models.Event{
		Type:      "PURCHASE_RECOMMENDATIONS",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"basket_id":       payload.BasketID,
			"terminal_id":     payload.TerminalID,
			"store_id":        payload.StoreID,
			"source_item_id":  payload.ItemID,
			"recommendations": recommendations,
		},
	}

	return []*models.Event{recommendEvent}, nil
}

func (p *Plugin) getRecommendations(ctx context.Context, itemID string) ([]map[string]interface{}, error) {
	rows, err := p.db.Pool().Query(ctx, `
		SELECT i.item_id, i.name, i.price, r.confidence_score
		FROM item_recommendations r
		JOIN items i ON i.item_id = r.recommended_item_id
		WHERE r.source_item_id = $1
		ORDER BY r.confidence_score DESC
		LIMIT 5
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query recommendations: %v", err)
	}
	defer rows.Close()

	var recommendations []map[string]interface{}
	for rows.Next() {
		var rec struct {
			ItemID          string  `json:"item_id"`
			Name            string  `json:"name"`
			Price           float64 `json:"price"`
			ConfidenceScore float64 `json:"confidence_score"`
		}

		if err := rows.Scan(&rec.ItemID, &rec.Name, &rec.Price, &rec.ConfidenceScore); err != nil {
			return nil, fmt.Errorf("failed to scan recommendation: %v", err)
		}

		recommendations = append(recommendations, map[string]interface{}{
			"item_id":          rec.ItemID,
			"name":             rec.Name,
			"price":            rec.Price,
			"confidence_score": rec.ConfidenceScore,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recommendations: %v", err)
	}

	return recommendations, nil
}

// UpdateRecommendations updates item recommendations based on purchase patterns
func (p *Plugin) UpdateRecommendations(ctx context.Context) error {
	_, err := p.db.Pool().Exec(ctx, `
		WITH basket_pairs AS (
			-- Find items that are frequently bought together
			SELECT 
				bi1.item_id as item1,
				bi2.item_id as item2,
				COUNT(*) as pair_count,
				COUNT(*) * 1.0 / (
					SELECT COUNT(*) 
					FROM basket_items 
					WHERE item_id = bi1.item_id
				) as confidence
			FROM basket_items bi1
			JOIN basket_items bi2 ON bi1.basket_id = bi2.basket_id
			WHERE bi1.item_id < bi2.item_id
			GROUP BY bi1.item_id, bi2.item_id
			HAVING COUNT(*) >= 2
		)
		INSERT INTO item_recommendations 
			(source_item_id, recommended_item_id, confidence_score)
		SELECT 
			item1, item2, confidence
		FROM basket_pairs
		WHERE confidence >= 0.1
		ON CONFLICT (source_item_id, recommended_item_id) 
		DO UPDATE SET 
			confidence_score = EXCLUDED.confidence_score,
			updated_at = CURRENT_TIMESTAMP
	`)

	if err != nil {
		return fmt.Errorf("failed to update recommendations: %v", err)
	}

	return nil
}
