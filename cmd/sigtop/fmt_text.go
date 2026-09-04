// Copyright (c) 2021, 2023 Tim van der Molen <tim@kariliq.nl>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/tbvdm/sigtop/signal"
)

func textWriteMessages(bw *bufio.Writer, msgs []signal.Message) {
	textWriteRecipientField(bw, "", "Conversation", msgs[0].Conversation)
	fmt.Fprintln(bw)
	for _, msg := range msgs {
		textWriteMessage(bw, &msg)
	}
}

func textWriteMessage(bw *bufio.Writer, msg *signal.Message) {
	if msg.IsOutgoing() {
		textWriteField(bw, "", "From", "You")
	} else if msg.Source != nil {
		textWriteRecipientField(bw, "", "From", msg.Source)
	}
	if msg.Type != "" {
		textWriteField(bw, "", "Type", msg.Type)
	} else {
		textWriteField(bw, "", "Type", "unknown")
	}
	if msg.TimeSent != 0 {
		textWriteTimeField(bw, "", "Sent", msg.TimeSent)
	}
	if !msg.IsOutgoing() {
		textWriteTimeField(bw, "", "Received", msg.TimeRecv)
	}
	textWriteAttachmentFields(bw, "", msg.Attachments)
	for _, rct := range msg.Reactions {
		textWriteFieldf(bw, "", "Reaction", "%s from %s", rct.Emoji, rct.Recipient.DetailedDisplayName())
	}
	if len(msg.Edits) == 0 {
		textWriteQuote(bw, "", msg.Quote)
		textWriteBody(bw, "", &msg.Body)
	} else {
		textWriteFieldf(bw, "", "Edited", "%d versions", len(msg.Edits))
		textWriteEditHistory(bw, msg.Edits)
	}
	fmt.Fprintln(bw)
}

func textWriteField(bw *bufio.Writer, prefix, field, value string) {
	if prefix != "" {
		prefix += " "
	}
	fmt.Fprintf(bw, "%s%s: %s\n", prefix, field, value)
}

func textWriteFieldf(bw *bufio.Writer, prefix, field, format string, a ...any) {
	textWriteField(bw, prefix, field, fmt.Sprintf(format, a...))
}

func textWriteRecipientField(bw *bufio.Writer, prefix, field string, rpt *signal.Recipient) {
	textWriteField(bw, prefix, field, rpt.DetailedDisplayName())
}

func textWriteTimeField(bw *bufio.Writer, prefix, field string, msec int64) {
	s := "unknown"
	if msec >= 0 {
		s = time.UnixMilli(msec).Format("Mon, 2 Jan 2006 15:04:05 -0700")
	}
	textWriteField(bw, prefix, field, s)
}

func textWriteAttachmentFields(bw *bufio.Writer, prefix string, atts []signal.Attachment) {
	for _, att := range atts {
		fileName := "no filename"
		if att.FileName != "" {
			fileName = att.FileName
		}
		textWriteFieldf(bw, prefix, "Attachment", "%s (%s, %d bytes)", fileName, att.ContentType, att.Size)
	}
}

func textWriteBody(bw *bufio.Writer, prefix string, body *signal.MessageBody) {
	if body.Text == "" {
		return
	}
	fmt.Fprintln(bw, prefix)
	if prefix != "" {
		prefix += " "
	}
	for _, line := range strings.Split(body.Text, "\n") {
		fmt.Fprintln(bw, prefix+line)
	}
}

func textWriteQuote(bw *bufio.Writer, prefix string, qte *signal.Quote) {
	if qte == nil {
		return
	}
	fmt.Fprintln(bw, prefix)
	if prefix != "" {
		prefix += " "
	}
	prefix += ">"
	textWriteRecipientField(bw, prefix, "From", qte.Recipient)
	textWriteTimeField(bw, prefix, "Sent", qte.TimeSent)
	textWriteQuoteAttachmentFields(bw, prefix, qte.Attachments)
	textWriteBody(bw, prefix, &qte.Body)
}

func textWriteQuoteAttachmentFields(bw *bufio.Writer, prefix string, atts []signal.QuoteAttachment) {
	for _, att := range atts {
		fileName := "no filename"
		if att.FileName != "" {
			fileName = att.FileName
		}
		textWriteFieldf(bw, prefix, "Attachment", "%s (%s)", fileName, att.ContentType)
	}
}

func textWriteEditHistory(bw *bufio.Writer, edits []signal.Edit) {
	fmt.Fprintln(bw)
	prefix := "|"
	for i := range edits {
		textWriteFieldf(bw, prefix, "Version", "%d", len(edits)-i)
		textWriteAttachmentFields(bw, prefix, edits[i].Attachments)
		textWriteTimeField(bw, prefix, "Sent", edits[i].TimeEdit)
		textWriteQuote(bw, prefix, edits[i].Quote)
		textWriteBody(bw, prefix, &edits[i].Body)
		if i+1 < len(edits) {
			fmt.Fprintln(bw, prefix)
		}
	}
}
