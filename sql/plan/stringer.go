// Copyright 2015 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package plan

import (
	"fmt"
	"strings"
)

// ToString explains a Plan, returns description string.
func ToString(p Plan) string {
	strs, _ := toString(p, []string{}, []int{})
	return strings.Join(strs, "->")
}

func toString(in Plan, strs []string, idxs []int) ([]string, []int) {
	switch in.(type) {
	case *Join, *Union, *Apply:
		idxs = append(idxs, len(strs))
	}

	for _, c := range in.GetChildren() {
		strs, idxs = toString(c, strs, idxs)
	}

	var str string
	switch x := in.(type) {
	case *Apply:
		last := len(idxs) - 1
		idx := idxs[last]
		children := strs[idx:]
		strs = strs[:idx]
		idxs = idxs[:last]
		str = "Apply{" + strings.Join(children, "->") + "}"
	case *Exists:
		str = "Exists"
	case *MaxOneRow:
		str = "MaxOneRow"
	case *Limit:
		str = "Limit"
	case *Sort:
		str = "Sort"
		if x.ExecLimit != nil {
			str += fmt.Sprintf(" + Limit(%v) + Offset(%v)", x.ExecLimit.Count, x.ExecLimit.Offset)
		}
	case *Join:
		last := len(idxs) - 1
		idx := idxs[last]
		children := strs[idx:]
		strs = strs[:idx]
		str = "Join{" + strings.Join(children, "->") + "}"
		idxs = idxs[:last]
		for _, eq := range x.EqualConditions {
			l := eq.GetArgs()[0].String()
			r := eq.GetArgs()[1].String()
			str += fmt.Sprintf("(%s,%s)", l, r)
		}
	case *Union:
		last := len(idxs) - 1
		idx := idxs[last]
		children := strs[idx:]
		strs = strs[:idx]
		str = "UnionAll{" + strings.Join(children, "->") + "}"
		idxs = idxs[:last]
	case *DataSource:
		if x.TableAsName != nil && x.TableAsName.L != "" {
			str = fmt.Sprintf("DataScan(%s)", x.TableAsName)
		} else {
			str = fmt.Sprintf("DataScan(%s)", x.tableInfo.Name)
		}
	case *Selection:
		str = "Selection"
	case *Projection:
		str = "Projection"
	case *Aggregation:
		str = "Aggr("
		for i, aggFunc := range x.AggFuncs {
			str += aggFunc.String()
			if i != len(x.AggFuncs)-1 {
				str += ","
			}
		}
		str += ")"
	case *Trim:
		str = "Trim"
	default:
		str = fmt.Sprintf("%T", in)
	}
	strs = append(strs, str)
	return strs, idxs
}
