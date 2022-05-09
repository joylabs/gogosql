package kipsql

import (
  "cloud.google.com/go/spanner"
  "context"
  "fmt"
  "time"
)

type SpannerTX interface {
	ReadRow(context.Context, string, spanner.Key, []string) (*spanner.Row, error)
}

// Column Definitions
type GoGoMessagesColumn struct {
    UserId string
    MessageId string
    Content string
    CreatedAt string
    UpdatedAt string
    IsDeleted string
    All []string
}
type GoGoCommunitiesColumn struct {
    CommunityId string
    VoteCount string
    Description string
    All []string
}

// Table Definitions
type GoGoMessagesTable struct {
    Columns GoGoMessagesColumn
    TableName string
}
type GoGoCommunitiesTable struct {
    Columns GoGoCommunitiesColumn
    TableName string
}

// All Table Definitions
type GoGoTables struct {
    Messages GoGoMessagesTable
    Communities GoGoCommunitiesTable
}

// PrimaryKey Definitions
type GoGoMessagesPrimaryKey struct {
    UserId string
    MessageId string
}
type GoGoCommunitiesPrimaryKey struct {
    CommunityId string
}

// Record Definitions
type Message struct {
    UserId string `spanner:"user_id"`
    MessageId string `spanner:"message_id"`
    Content string `spanner:"content"`
    created_at time.Time `spanner:"created_at"`
    UpdatedAt time.Time `spanner:"updated_at"`
    IsDeleted bool `spanner:"is_deleted"`
}
type PartialMessage struct {
    UserId *string
    MessageId *string
    Content *string
    created_at *time.Time
    UpdatedAt *time.Time
    IsDeleted *bool
}
type Comm struct {
    CommunityId string `spanner:"community_id"`
    VoteCount int `spanner:"vote_count"`
    Description string `spanner:"description"`
}
type PartialComm struct {
    CommunityId *string
    VoteCount *int
    Description *string
}

var (
    Tables = GoGoTables{
        Messages: GoGoMessagesTable{
            Columns: GoGoMessagesColumn{
                UserId: "user_id",
                MessageId: "message_id",
                Content: "content",
                CreatedAt: "created_at",
                UpdatedAt: "updated_at",
                IsDeleted: "is_deleted",
                All: []string{
                    "user_id",
                    "message_id",
                    "content",
                    "created_at",
                    "updated_at",
                    "is_deleted",
                },
            },
            TableName: "messages",
        },
        Communities: GoGoCommunitiesTable{
            Columns: GoGoCommunitiesColumn{
                CommunityId: "community_id",
                VoteCount: "vote_count",
                Description: "description",
                All: []string{
                    "community_id",
                    "vote_count",
                    "description",
                },
            },
            TableName: "communities",
        },
  }
)

// Methods
func (GoGoMessagesTable) PrimaryKey (
    userId string,
    messageId string,
) GoGoMessagesPrimaryKey {
	return GoGoMessagesPrimaryKey{
            userId,
            messageId,
        }
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

func (r Message) PrimaryKey() GoGoMessagesPrimaryKey {
	return GoGoMessagesPrimaryKey{
            r.UserId,
            r.MessageId,
        }
}

func (r PartialMessage) PrimaryKey() (GoGoMessagesPrimaryKey, error) {
    if r.UserId == nil {
        return GoGoMessagesPrimaryKey{}, fmt.Errorf("UserId is required to make a spanner key for: %v", r)
    }
    if r.MessageId == nil {
        return GoGoMessagesPrimaryKey{}, fmt.Errorf("MessageId is required to make a spanner key for: %v", r)
    }

    return GoGoMessagesPrimaryKey{
            *r.UserId,
            *r.MessageId,
        }, nil
    }

func (r GoGoMessagesPrimaryKey) SpannerKey() spanner.Key {
	return spanner.Key{
            r.UserId,
            r.MessageId,
        }
}

func (GoGoCommunitiesTable) PrimaryKey (
    communityId string,
) GoGoCommunitiesPrimaryKey {
	return GoGoCommunitiesPrimaryKey{
            communityId,
        }
}

func (t GoGoCommunitiesTable) ReadRow(ctx context.Context, tx SpannerTX, key spanner.Key) (*Comm, error) {
	row, err := tx.ReadRow(ctx, t.TableName, key, t.Columns.All)
	if err != nil {
		return nil, err
	}

	r := &Comm{}
	err = row.ToStruct(r)
	return r, err
}

func (r Comm) PrimaryKey() GoGoCommunitiesPrimaryKey {
	return GoGoCommunitiesPrimaryKey{
            r.CommunityId,
        }
}

func (r PartialComm) PrimaryKey() (GoGoCommunitiesPrimaryKey, error) {
    if r.CommunityId == nil {
        return GoGoCommunitiesPrimaryKey{}, fmt.Errorf("CommunityId is required to make a spanner key for: %v", r)
    }

    return GoGoCommunitiesPrimaryKey{
            *r.CommunityId,
        }, nil
    }

func (r GoGoCommunitiesPrimaryKey) SpannerKey() spanner.Key {
	return spanner.Key{
            r.CommunityId,
        }
}
