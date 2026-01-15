package anthropic

const (
	MessageContentToolResultType = "tool_result"
	MessageContentTextType       = "text"
	MessageContentImageType      = "image"
	MessageContentImageURLType   = "image_url"
	MessageContentVideoType      = "video"
	MessageContentAudioType      = "audio"
	MessageContentFileType       = "file"
	MessageContentLinkType       = "link"
	MessageContentDocumentType   = "document"
	MessageContentCodeType       = "code"
)

const (
	MessageParamRoleUser      = "user"
	MessageParamRoleAssistant = "assistant"
	MessageParamRoleSystem    = "system"
	MessageParamRoleTool      = "tool"
)
