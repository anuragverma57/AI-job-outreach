package model

type SmartApplyRequest struct {
	RawText string `json:"raw_text"`
}

type SmartApplyMeta struct {
	ExtractionConfidence string `json:"extraction_confidence"`
}

type SmartApplyResponse struct {
	Application *Application   `json:"application"`
	Email       *Email         `json:"email"`
	Meta        SmartApplyMeta `json:"meta"`
}
