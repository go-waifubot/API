package db

import (
	"context"
	"errors"
	"time"
)

type Querier interface {
	Profile(context.Context, int64) (*Profile, error)
	UserByAnilistURL(context.Context, string) (*User, error)
}

type Profile struct {
	ID         int64  `json:"id,string"`
	Quote      string `json:"quote,omitempty"`
	Tokens     int32  `json:"tokens,omitempty"`
	AnilistURL string `json:"anilist_url,omitempty"`
	Favorite   Char   `json:"favorite,omitempty"`
	Waifus     []Char `json:"waifus,omitempty"`
}

type Char struct {
	Date  time.Time `json:"date"`
	Name  string    `json:"name"`
	Image string    `json:"image"`
	Type  string    `json:"type"`
	ID    int64     `json:"id"`
}

func (q *Queries) Profile(ctx context.Context, userID int64) (*Profile, error) {
	p, err := q.getProfile(ctx, userID)
	if err != nil {
		return nil, err
	} else if len(p) == 0 {
		return nil, errors.New("nothing found")
	}

	return mapUser(p...), nil
}

func (q *Queries) UserByAnilistURL(ctx context.Context, anilistURL string) (*User, error) {
	u, err := q.getUserByAnilist(ctx, anilistURL)
	if err != nil {
		return nil, err
	}

	return &User{
		UserID:     u.UserID,
		Quote:      u.Quote,
		Date:       u.RollDate,
		Favorite:   u.Favorite,
		Tokens:     u.Tokens,
		AnilistUrl: u.AnilistUrl,
	}, nil
}

func mapUser(userRows ...getProfileRow) *Profile {
	if len(userRows) < 1 {
		return nil
	}

	p := &Profile{
		ID:         userRows[0].UserID,
		Quote:      userRows[0].Quote,
		Tokens:     userRows[0].Tokens,
		AnilistURL: userRows[0].AnilistUrl,
		Waifus:     make([]Char, 0, len(userRows)),
	}

	for _, u := range userRows {
		if u.Favorite.Int64 == u.ID {
			p.Favorite = Char{
				ID:    u.ID,
				Name:  u.Name,
				Image: u.Image,
				Type:  u.Type,
				Date:  u.Date,
			}
		}

		p.Waifus = append(p.Waifus, Char{
			ID:    u.ID,
			Name:  u.Name,
			Image: u.Image,
			Type:  u.Type,
			Date:  u.Date,
		})
	}

	return p
}
