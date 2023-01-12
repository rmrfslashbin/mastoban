package app

// Errors for the app module.

func errorFailedToSuspendUser() string {
	msg := "failed to suspend user"
	return msg
}

func errorMessageEventNotSupported() string {
	msg := "message event not supported"
	return msg
}

func errorMissingPSK() string {
	msg := "missing psk query param"
	return msg
}

func errorPSKMismatch() string {
	msg := "provided PSK is invalid"
	return msg
}

func errorUnableToCreateGeoIPInstance() string {
	msg := "unable to create GeoIP instance"
	return msg
}

func errorUnableToCreateMastoclientInstance() string {
	msg := "unable to create mastoclient instance"
	return msg
}

func errorUnableToCreateQueueInstance() string {
	msg := "unable to create queue instance"
	return msg
}

func errorUnableToFetchEnvVar(varname string) string {
	msg := "unable to fetch environment variable: " + varname
	return msg
}

func errorUnableToLookupIP() string {
	msg := "unable to lookup IP in GeoIP database"
	return msg
}

func errorUnableToSendMessageToQueue() string {
	msg := "unable to send message to queue"
	return msg
}

func errorUnableToUnmarshalRequest() string {
	msg := "unable to unmarshal request"
	return msg
}
