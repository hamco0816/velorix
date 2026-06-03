//go:build unit

package service

import (
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// 构建的带附件邮件应为合法的 multipart/mixed，附件经 base64 编码后可还原。
func TestBuildMultipartMessage_AttachmentRoundTrip(t *testing.T) {
	attachment := []byte("%PDF-1.4 fake invoice content with binary \x00\x01\x02 bytes")
	msg := buildMultipartMessage(
		"noreply@example.com",
		"customer@example.com",
		"您申请的发票已开具",
		"sub2api-test-boundary",
		"<p>发票正文</p>",
		"invoice-123.pdf",
		"application/pdf",
		attachment,
	)
	raw := string(msg)

	// 头部：multipart/mixed + 主题 MIME 编码
	require.Contains(t, raw, `Content-Type: multipart/mixed; boundary="sub2api-test-boundary"`)
	require.Contains(t, raw, "Subject: "+mime.QEncoding.Encode("UTF-8", "您申请的发票已开具"))
	require.Contains(t, raw, `Content-Disposition: attachment; filename="invoice-123.pdf"`)

	// 用标准库解析 multipart，校验两个分段并还原附件
	headerEnd := strings.Index(raw, "\r\n\r\n")
	require.Positive(t, headerEnd)
	reader := multipart.NewReader(strings.NewReader(raw[headerEnd+4:]), "sub2api-test-boundary")

	htmlPart, err := reader.NextPart()
	require.NoError(t, err)
	htmlBody, err := io.ReadAll(htmlPart)
	require.NoError(t, err)
	require.Equal(t, "<p>发票正文</p>", strings.TrimSpace(string(htmlBody)))

	attachPart, err := reader.NextPart()
	require.NoError(t, err)
	require.Equal(t, "base64", attachPart.Header.Get("Content-Transfer-Encoding"))
	encoded, err := io.ReadAll(attachPart)
	require.NoError(t, err)
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(strings.TrimSpace(string(encoded)), "\r\n", ""))
	require.NoError(t, err)
	require.Equal(t, attachment, decoded)
}

// base64 行编码应按 76 字符换行（RFC 2045）。
func TestEncodeBase64Lines_Wrapping(t *testing.T) {
	data := make([]byte, 200)
	out := encodeBase64Lines(data)
	for _, line := range strings.Split(strings.TrimRight(out, "\r\n"), "\r\n") {
		require.LessOrEqual(t, len(line), 76)
	}
}
