# Payment Expiration Handling

## Overview
This document explains how expired payments are handled in the MongoCollectibles rental platform, specifically addressing the issue where PayMongo redirect URLs don't trigger automatically when payments expire.

## The Problem

When using PayMongo's payment sources (GCash, GrabPay, bank transfers), the redirect URLs only work when:
- User visits the checkout URL and completes the payment → redirects to success URL
- User visits the checkout URL and cancels → redirects to failed URL

**However**, PayMongo does NOT automatically redirect when:
- The payment source expires without being visited
- The timer runs out before user completes payment
- The source status changes to "expired" on PayMongo's side

## The Solution

We've implemented a **multi-layer approach** to handle payment expiration:

### 1. Backend Auto-Expiration (Server-Side)
**File**: `internal/service/rental.go`
**Lines**: 129-163 (GetUserRentals function)

When a user loads their rentals dashboard, the backend automatically:
- Checks all `pending_payment` rentals
- Calculates time remaining based on `expires_at` timestamp
- Expires any rentals where `time_remaining <= 0`
- Returns units to available inventory
- Updates rental status to `expired`

```go
// Check if rental has expired
if rental.Status == "pending_payment" && rental.ExpiresAt != nil {
    if time.Now().After(*rental.ExpiresAt) {
        // Expire and return units
        s.ExpireRental(rental.ID, tx)
    }
}
```

### 2. Status Polling (Client-Side - Payment Page)
**File**: `web/payment.html`
**Lines**: 446 (statusCheckInterval), 516-527 (startStatusPolling function)

The payment page now polls the backend every 5 seconds to check rental status:
- If status changes from `pending_payment` to `expired` → redirects to `/payment-failed?reason=expired`
- If status changes from `pending_payment` to `active` → redirects to `/payment-success?rental_id=X`

This ensures users are automatically redirected even if PayMongo's redirect fails.

```javascript
startStatusPolling() {
  this.statusCheckInterval = setInterval(async () => {
    const response = await fetch(`/api/rentals/${this.rental.id}`);
    const data = await response.json();
    
    if (data.rental.status === 'expired') {
      window.location.href = '/payment-failed?reason=expired';
    } else if (data.rental.status === 'active') {
      window.location.href = '/payment-success?rental_id=' + this.rental.id;
    }
  }, 5000); // Check every 5 seconds
}
```

### 3. Success Page Verification (Client-Side)
**File**: `web/payment-success.html`
**Lines**: 104-139 (verifyPayment function)

When PayMongo redirects to the success page, we verify the actual payment status:
- Calls `/api/payments/:id/verify` to check real status
- If status is `failed`, `expired`, or `cancelled` → redirects to failed page
- Only shows success if status is actually `paid`

This handles the edge case where PayMongo redirects to success URL but payment actually failed.

```javascript
async function verifyPayment() {
  const response = await fetch(`/api/payments/${paymentId}/verify`);
  const data = await response.json();
  
  // Check if payment actually failed/expired/cancelled
  if (data.status === 'failed' || data.status === 'expired' || 
      data.status === 'cancelled' || !data.success) {
    window.location.href = '/payment-failed?payment_id=' + paymentId;
    return;
  }
  
  showSuccessUI(data);
}
```

### 4. Failed Page Verification (Client-Side)
**File**: `web/payment-failed.html`
**Lines**: 42-83 (verification script)

When PayMongo redirects to the failed page, we verify the actual payment status:
- Calls `/api/payments/:id/verify` to check real status
- If status is actually `paid` → redirects to success page
- Shows appropriate failure reason based on status (expired, cancelled, failed)

This handles the edge case where PayMongo redirects to failed URL but payment actually succeeded.

```javascript
async function verifyPaymentStatus() {
  const response = await fetch(`/api/payments/${paymentId}/verify`);
  const data = await response.json();
  
  // If payment actually succeeded, redirect to success page
  if (data.success && data.status === 'paid') {
    window.location.href = '/payment-success?payment_id=' + paymentId;
    return;
  }
  
  showFailedUI(data.status);
}
```

## Flow Diagram

```
User creates rental → Status: pending_payment
                     Timer starts: 1 minute (configurable)
                     
                     ↓
                     
Payment Page → Every 5 seconds: Check rental status
               - If expired → Redirect to failed page
               - If active → Redirect to success page
               
               User clicks "Pay with GCash"
                     ↓
                     
PayMongo Checkout → User completes → Redirect to success page
                   User cancels → Redirect to failed page
                   User leaves → No redirect (handled by polling)
                   Timer expires → No redirect (handled by polling)
                   
                     ↓
                     
Success/Failed Page → Verify actual payment status
                     - If mismatch → Redirect to correct page
                     - If match → Show appropriate message
```

## Timer Configuration

**File**: `internal/service/rental.go`
**Lines**: 76-78

```go
// Examples: 30 * time.Second, 1 * time.Minute, 5 * time.Minute, 10 * time.Minute
expiresAt := time.Now().Add(1 * time.Minute) // Currently 1 minute for testing
```

**For Production**: Change to `5 * time.Minute` or `10 * time.Minute`

## Testing the Feature

1. Create a rental and proceed to payment
2. Select GCash or any e-wallet method
3. **Do NOT visit the checkout URL** - just wait on the payment page
4. Watch the countdown timer reach 0
5. Within 5 seconds after expiration, you should be automatically redirected to the failed page
6. The rental status will be `expired` and units returned to inventory

## Key Benefits

✅ **No Manual Refresh Required**: Status polling automatically detects changes
✅ **Handles PayMongo Redirect Failures**: Verification catches status mismatches
✅ **Units Always Return**: Backend expiration ensures inventory is freed
✅ **User-Friendly**: Automatic redirects with clear messaging
✅ **Configurable Timer**: Easy to adjust for testing vs production

## Future Enhancements

### PayMongo Webhooks (Recommended)
For production environments, consider implementing PayMongo webhooks:
- More reliable than polling
- Real-time status updates
- Reduces API calls
- Better scalability

**Webhook Events to Handle**:
- `source.chargeable` - Payment completed successfully
- `payment.paid` - Payment confirmed
- `payment.failed` - Payment failed

**Implementation**:
1. Create webhook endpoint: `POST /api/webhooks/paymongo`
2. Verify webhook signature for security
3. Update rental and payment status based on event
4. Return units if payment failed/expired
