package constants

// Error code
const (
	SYSTEM_ERROR      = "system_error"      //
	REQUEST_INVALID   = "request_invalid"   //
	RECORD_NOT_EXIST  = "record_not_exist"  //
	RECORD_EXISTED    = "record_existed"    //
	PERMISSION_DENIED = "permission_denied" //
	TOKEN_MISSING     = "token_missing"     //
	STATUS_CONFLICT   = "status_conflict"   //
	INVALID_FORMAT    = "invalid_format"    //
	ERROR_MAP_DATA    = "error_map_data"    //
)

// Error message
const (
	SystemErrorMessage    = "There was an error on the server side"
	RequestInvalidMessage = "Invalid request"
	RecordNotExistMessage = "data does not exist"
	RecordExistMessage    = "data already exists"
	StatusConflictMessage = "The record has been modified by another process. Please refresh and try again."
)
