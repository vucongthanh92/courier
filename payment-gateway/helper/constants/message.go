package constants

// Error system code
const (
	SYSTEM_ERROR        = "system_error"
	REQUEST_INVALID     = "request_invalid"
	RECORD_NOT_EXIST    = "record_not_exist"
	RECORD_EXISTED      = "record_existed"
	PERMISSION_DENIED   = "permission_denied"
	TOKEN_MISSING       = "token_missing"
	STATUS_CONFLICT     = "status_conflict"
	INVALID_FORMAT      = "invalid_format"
	ERROR_MAP_DATA      = "error_map_data"
	USER_ALREADY_EXISTS = "user_already_exists"
)

// Error message
const (
	// general
	SystemErrorMessage    = "There was an error on the server side"
	RequestInvalidMessage = "Invalid request"
	RecordNotExistMessage = "data does not exist"
	RecordExistMessage    = "data already exists"
	StatusConflictMessage = "The record has been modified by another process. Please refresh and try again."

	// new
	ErrIdempotencyConflict         = "idempotency key is already processing or was used with a different request"
	ErrWalletInactive              = "wallet is not active"
	ErrProviderSepayCheckoutFailed = "cannot create SePay checkout"
)

// Error payment code
const (
	WALLET_INACTIVE_CODE          = "wallet_inactive"
	PROVIDER_CHECKOUT_FAILED_CODE = "provider_checkout_failed" // This code is used when the payment gateway fails to create a checkout with the provider (e.g., SePay).
	TRANSACTION_NOT_ELIGIBLE_CODE = "transaction_not_eligible" // This code is used when a transaction does not meet the criteria for processing, such as an invalid transfer type or missing payment code.
	TOPUP_INTENT_NOT_FOUND_CODE   = "topup_intent_not_found"   // This code is used when a top-up intent associated with a payment code cannot be found in the system.
	TOPUP_ALREADY_SUCCEEDED_CODE  = "topup_already_succeeded"  // This code is used when a top-up intent has already been marked as succeeded.
	AMOUNT_MISMATCH_CODE          = "amount_mismatch"          // This code is used when the transfer amount in the webhook does not match the expected amount in the top-up intent.
)

// Information message
const (
	CourierWalletTopUpDescription = "Courier wallet top-up"
)
