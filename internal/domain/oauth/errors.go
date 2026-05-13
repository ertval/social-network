package oauth

import "errors"

var (
	ErrUserNotFound                        = errors.New("oauth user not found")
	ErrUserWithEmailExists                 = errors.New("oauth user not found")
	ErrProviderAccountBelongsToAnotherUser = errors.New("Provider Account Belongs To Another User")
	ErrAlreadyLinkedToProvider             = errors.New("User Already Linked to Provider")
)
