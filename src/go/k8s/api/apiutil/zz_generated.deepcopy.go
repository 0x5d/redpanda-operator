//go:build !ignore_autogenerated

// Copyright 2022 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

// Code generated by controller-gen. DO NOT EDIT.

package apiutil

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JSONBoolean) DeepCopyInto(out *JSONBoolean) {
	*out = *in
	if in.Raw != nil {
		in, out := &in.Raw, &out.Raw
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JSONBoolean.
func (in *JSONBoolean) DeepCopy() *JSONBoolean {
	if in == nil {
		return nil
	}
	out := new(JSONBoolean)
	in.DeepCopyInto(out)
	return out
}
