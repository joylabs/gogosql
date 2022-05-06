package gogosql

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
)

type SpannerTX interface {
	ReadRow(context.Context, string, spanner.Key, []string) (*spanner.Row, error)
}

// TableDefinitions
type GoGoMessagesColumns struct {
	UserId    string
	MessageId string
	Content   string
	CreatedAt string
	All       []string
}

type GoGoMessagesTable struct {
	Columns   GoGoMessagesColumns
	TableName string
}

type GoGoTables struct {
	Messages GoGoMessagesTable
}

// RecordDefinitions
type Message struct {
	UserId    string    `spanner:"user_id"`
	MessageId string    `spanner:"message_id"`
	Content   string    `spanner:"content"`
	CreatedAt time.Time `spanner:"created_at"`
}

type MessageOpt struct {
	UserId    *string
	MessageId *string
	Content   *string
	CreatedAt *time.Time
}

var (
	KipTables = GoGoTables{
		Messages: GoGoMessagesTable{
			Columns: GoGoMessagesColumns{
				UserId:    "user_id",
				MessageId: "message_id",
				Content:   "content",
				CreatedAt: "created_at",
				All: []string{
					"user_id",
					"message_id",
					"content",
					"created_at",
				},
			},
		},
	}
)

func (GoGoMessagesTable) NewSpannerKey(userId string, messageId string) spanner.Key {
	return spanner.Key{userId, messageId}
}

func (t GoGoMessagesTable) ReadRow(ctx context.Context, tx SpannerTX, key spanner.Key) (*Message, error) {
	row, err := tx.ReadRow(ctx, t.TableName, key, t.Columns.All)
	if err != nil {
		return nil, err
	}

	r := &Message{}
	err = row.ToStruct(r)
	return r, err
}

func (r Message) SpannerKey() spanner.Key {
	return spanner.Key{r.UserId, r.MessageId}
}

func (r MessageOpt) SpannerKey() (spanner.Key, error) {
	if r.UserId == nil {
		return nil, fmt.Errorf("UserId is required to make a spanner key for: %v", r)
	}

	if r.MessageId == nil {
		return nil, fmt.Errorf("MessageId is required to make a spanner key for: %v", r)
	}

	return spanner.Key{*r.UserId, *r.MessageId}, nil
}
