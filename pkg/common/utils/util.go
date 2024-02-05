// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

// Union 获取切片中所有元素的并集（去重）
func Union[E comparable](es ...[]E) []E {
	if len(es) == 0 {
		return []E{}
	}

	// 创建一个映射，用于记录元素是否已经存在
	seen := make(map[E]struct{})
	var result []E

	// 遍历每个切片
	for _, slice := range es {
		// 遍历切片中的元素
		for _, elem := range slice {
			// 如果元素未在映射中出现，则加入结果切片，并标记为已出现
			if _, ok := seen[elem]; !ok {
				result = append(result, elem)
				seen[elem] = struct{}{}
			}
		}
	}

	return result
}
