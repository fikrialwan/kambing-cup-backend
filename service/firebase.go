package service

import (
	"context"

	"firebase.google.com/go/v4/db"
)

// Firebase Interfaces
type FirebaseClient interface {
	NewRef(path string) FirebaseRef
}

type FirebaseRef interface {
	Set(ctx context.Context, v interface{}) error
}

// Real Implementation
type RealFirebaseClient struct {
	Client *db.Client
}

func (c *RealFirebaseClient) NewRef(path string) FirebaseRef {
	return &RealFirebaseRef{Ref: c.Client.NewRef(path)}
}

type RealFirebaseRef struct {
	Ref *db.Ref
}

func (r *RealFirebaseRef) Set(ctx context.Context, v interface{}) error {
	return r.Ref.Set(ctx, v)
}
