/*
 * Copyright 2018 Google Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package explore provides a high-performance table-based implementation of the
// ExploreService defined in kythe/proto/explore.proto.
//
// Table format:
//   <parent ticket>   -> epb.Relatives
//   <ticket>          -> epb.Relatives
package explore

import (
	"context"
	"fmt"

	"kythe.io/kythe/go/storage/table"

	epb "kythe.io/kythe/proto/explore_go_proto"
	srvpb "kythe.io/kythe/proto/serving_go_proto"
)

// Tables implements the explore.Service interface using separate static lookup tables
// for each API component.
type Tables struct {
	// ParentToChildren is a table of srvpb.Relatives keyed by parent ticket.
	ParentToChildren table.Proto

	// ChildToParents is a table of srvpb.Relatives keyed by child ticket.
	ChildToParents table.Proto
}

// TypeHierarchy returns the hierarchy (supertypes and subtypes, including implementations)
// of a specified type, as a directed acyclic graph.
// TODO: not yet implemented
func (t *Tables) TypeHierarchy(ctx context.Context, req *epb.TypeHierarchyRequest) (*epb.TypeHierarchyReply, error) {
	return nil, nil
}

// Callers returns the (recursive) callers of a specified function, as a directed graph.
// TODO: not yet implemented
func (t *Tables) Callers(ctx context.Context, req *epb.CallersRequest) (*epb.CallersReply, error) {
	return nil, nil
}

// Callees returns the (recursive) callees of a specified function
// (that is, what functions this function calls), as a directed graph.
// TODO: not yet implemented
func (t *Tables) Callees(ctx context.Context, req *epb.CalleesRequest) (*epb.CalleesReply, error) {
	return nil, nil
}

// Parameters returns the parameters of a specified function.
// TODO: not yet implemented
func (t *Tables) Parameters(ctx context.Context, req *epb.ParametersRequest) (*epb.ParametersReply, error) {
	return nil, nil
}

// Parents returns the parents of a specified node
// (for example, the file for a class, or the class for a function).
// Note: in some cases a node may have more than one parent.
// TODO: populate NodeData appropriately
func (t *Tables) Parents(ctx context.Context, req *epb.ParentsRequest) (*epb.ParentsReply, error) {
	childTickets := req.Tickets
	if len(childTickets) == 0 {
		return nil, fmt.Errorf("missing input tickets: %v", &req)
	}

	reply := &epb.ParentsReply{
		InputToParents: make(map[string]*epb.Tickets),
	}

	// At the moment, this is our policy for missing data: if a child (input) ticket has
	// (a) no record in the table, we don't include a mapping for that ticket in the response.
	// (b) an empty set of parents in the table, we include a mapping from that ticket to nil
	// Other table access errors result in returning an error.
	for _, ticket := range childTickets {
		var relatives srvpb.Relatives
		err := t.ChildToParents.Lookup(ctx, []byte(ticket), &relatives)

		if err != nil && err != table.ErrNoSuchKey {
			return nil, fmt.Errorf("error looking up parents with ticket %q: %v", ticket, err)
		}

		if relatives.Type != srvpb.Relatives_PARENTS {
			return nil, fmt.Errorf("type of relatives is not 'PARENTS': %v", relatives)
		}

		// TODO: consider logging a warning if len(relatives.Tickets) == 0
		// (postprocessing should disallow this)
		if len(relatives.Tickets) != 0 {
			reply.InputToParents[ticket] = &epb.Tickets{Tickets: relatives.Tickets}
		}
	}

	return reply, nil
}

// Children returns the children of a specified node
// (for example, the classes contained in a file, or the functions contained in a class).
// TODO: populate NodeData appropriately
func (t *Tables) Children(ctx context.Context, req *epb.ChildrenRequest) (*epb.ChildrenReply, error) {
	parentTickets := req.Tickets
	if len(parentTickets) == 0 {
		return nil, fmt.Errorf("missing input tickets: %v", &req)
	}

	reply := &epb.ChildrenReply{
		InputToChildren: make(map[string]*epb.Tickets),
	}

	// At the moment, this is our policy for missing data: if a parent (input) ticket has
	// (a) no record in the table, we don't include a mapping for that ticket in the response.
	// (b) an empty set of children in the table, we include a mapping from that ticket to nil
	// Other table access errors result in returning an error.
	for _, ticket := range parentTickets {
		var relatives srvpb.Relatives
		err := t.ParentToChildren.Lookup(ctx, []byte(ticket), &relatives)

		if err != nil && err != table.ErrNoSuchKey {
			return nil, fmt.Errorf("error looking up children with ticket %q: %v", ticket, err)
		}

		if relatives.Type != srvpb.Relatives_CHILDREN {
			return nil, fmt.Errorf("type of relatives is not 'CHILDREN': %v", relatives)
		}

		// TODO: consider logging a warning if len(relatives.Tickets) == 0
		// (postprocessing should disallow this)
		if len(relatives.Tickets) != 0 {
			reply.InputToChildren[ticket] = &epb.Tickets{Tickets: relatives.Tickets}
		}
	}

	return reply, nil
}
