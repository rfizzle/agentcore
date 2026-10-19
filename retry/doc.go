// Package retry is the bounded wait every driver of an agent reaches for when
// the provider never answered a request: how long to wait, how many times,
// and the notice a surface says about each attempt. The waiting itself is the
// driver's to do.
package retry
