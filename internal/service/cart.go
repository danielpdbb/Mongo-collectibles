package service

import (
	"github.com/danielpdbb/Mongo-collectibles/internal/domain"
	"github.com/danielpdbb/Mongo-collectibles/internal/repository"
)

// ========================================
// CART SERVICE
// ========================================

// GetOrCreateCart retrieves the user's cart or creates one if it doesn't exist
func GetOrCreateCart(userID uint) (*domain.Cart, error) {
	var cart domain.Cart

	// Try to find existing cart
	err := repository.DB.Where("user_id = ?", userID).First(&cart).Error

	if err != nil {
		// Cart doesn't exist, create one
		cart = domain.Cart{UserID: userID}
		if err := repository.DB.Create(&cart).Error; err != nil {
			return nil, err
		}
	}

	return &cart, nil
}

// GetCartWithItems retrieves the cart with all items and calculated prices
func GetCartWithItems(userID uint) (*domain.CartResponse, error) {
	cart, err := GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// Load cart items with relationships
	var items []domain.CartItem
	err = repository.DB.
		Preload("Collectible").
		Preload("Store").
		Where("cart_id = ?", cart.ID).
		Order("created_at DESC").
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	// Build response with calculated prices
	var itemResponses []domain.CartItemResponse
	totalItems := 0
	totalPrice := 0

	for _, item := range items {
		// Calculate price for this item
		unitPrice := CalculateRentalPrice(item.Collectible.Size, item.Days)
		itemTotal := unitPrice * item.Quantity

		// Get available units
		availableUnits := CountAvailableUnits(item.CollectibleID)

		itemResponses = append(itemResponses, domain.CartItemResponse{
			ID:             item.ID,
			CollectibleID:  item.CollectibleID,
			StoreID:        item.StoreID,
			Quantity:       item.Quantity,
			Days:           item.Days,
			UnitPrice:      unitPrice,
			ItemTotal:      itemTotal,
			Collectible:    item.Collectible,
			Store:          item.Store,
			AvailableUnits: availableUnits,
		})

		totalItems += item.Quantity
		totalPrice += itemTotal
	}

	return &domain.CartResponse{
		ID:         cart.ID,
		Items:      itemResponses,
		TotalItems: totalItems,
		TotalPrice: totalPrice,
		ItemCount:  len(items),
	}, nil
}

// AddToCart adds an item to the user's cart
// If the same collectible+store+days already exists, it updates the quantity
func AddToCart(userID uint, req domain.AddToCartRequest) (*domain.CartItem, error) {
	cart, err := GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// Check if collectible exists
	var collectible domain.Collectible
	if err := repository.DB.First(&collectible, req.CollectibleID).Error; err != nil {
		return nil, err
	}

	// Check if same item already in cart (same collectible, store, and days)
	var existingItem domain.CartItem
	err = repository.DB.
		Where("cart_id = ? AND collectible_id = ? AND store_id = ? AND days = ?",
			cart.ID, req.CollectibleID, req.StoreID, req.Days).
		First(&existingItem).Error

	if err == nil {
		// Item exists, update quantity
		newQuantity := existingItem.Quantity + req.Quantity

		// Check stock
		availableUnits := CountAvailableUnits(req.CollectibleID)
		if int64(newQuantity) > availableUnits {
			newQuantity = int(availableUnits) // Cap at available
		}

		existingItem.Quantity = newQuantity
		repository.DB.Save(&existingItem)

		// Load relationships
		repository.DB.Preload("Collectible").Preload("Store").First(&existingItem, existingItem.ID)
		return &existingItem, nil
	}

	// Check stock for new item
	availableUnits := CountAvailableUnits(req.CollectibleID)
	if int64(req.Quantity) > availableUnits {
		req.Quantity = int(availableUnits) // Cap at available
		if req.Quantity == 0 {
			return nil, nil // No stock available
		}
	}

	// Create new cart item
	item := domain.CartItem{
		CartID:        cart.ID,
		CollectibleID: req.CollectibleID,
		StoreID:       req.StoreID,
		Quantity:      req.Quantity,
		Days:          req.Days,
	}

	if err := repository.DB.Create(&item).Error; err != nil {
		return nil, err
	}

	// Load relationships for response
	repository.DB.Preload("Collectible").Preload("Store").First(&item, item.ID)

	return &item, nil
}

// UpdateCartItem updates quantity and days for a cart item
func UpdateCartItem(userID uint, itemID uint, req domain.UpdateCartItemRequest) (*domain.CartItem, error) {
	cart, err := GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	var item domain.CartItem
	err = repository.DB.
		Where("id = ? AND cart_id = ?", itemID, cart.ID).
		First(&item).Error

	if err != nil {
		return nil, err
	}

	// Check stock
	availableUnits := CountAvailableUnits(item.CollectibleID)
	if int64(req.Quantity) > availableUnits {
		req.Quantity = int(availableUnits)
	}

	// Update item
	item.Quantity = req.Quantity
	item.Days = req.Days
	repository.DB.Save(&item)

	// Load relationships
	repository.DB.Preload("Collectible").Preload("Store").First(&item, item.ID)

	return &item, nil
}

// RemoveFromCart removes an item from the user's cart
func RemoveFromCart(userID uint, itemID uint) error {
	cart, err := GetOrCreateCart(userID)
	if err != nil {
		return err
	}

	return repository.DB.
		Where("id = ? AND cart_id = ?", itemID, cart.ID).
		Delete(&domain.CartItem{}).Error
}

// ClearCart removes all items from the user's cart
func ClearCart(userID uint) error {
	cart, err := GetOrCreateCart(userID)
	if err != nil {
		return err
	}

	return repository.DB.
		Where("cart_id = ?", cart.ID).
		Delete(&domain.CartItem{}).Error
}

// GetCartItemCount returns the total number of items in cart
func GetCartItemCount(userID uint) int {
	cart, err := GetOrCreateCart(userID)
	if err != nil {
		return 0
	}

	var count int64
	repository.DB.Model(&domain.CartItem{}).
		Where("cart_id = ?", cart.ID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&count)

	return int(count)
}
