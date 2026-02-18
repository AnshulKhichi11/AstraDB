package engine

func DecodeRequest(body []byte, out any) error {
	return decodeJSONUseNumber(body, out)
}
