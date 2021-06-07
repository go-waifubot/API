package db

import (
	"context"
	"time"
)

type Querier interface {
	Profile(context.Context, int64) (*Profile, error)
}

type Profile struct {
	Quote    string `json:"quote,omitempty"`
	Waifus   []Char `json:"waifus,omitempty"`
	ID       int64  `json:"id"`
	Favorite int64  `json:"favorite,omitempty"`
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
	}

	return mapUser(p...), nil
}

func mapUser(userRows ...getProfileRow) *Profile {
	if len(userRows) < 1 {
		return nil
	}

	p := &Profile{
		ID:       userRows[0].UserID,
		Favorite: userRows[0].Favorite.Int64,
		Quote:    userRows[0].Quote,
		Waifus:   make([]Char, 0, len(userRows)),
	}

	for _, u := range userRows {
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
