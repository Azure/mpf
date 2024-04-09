//     MIT License
// 
//     Copyright (c) Microsoft Corporation.
// 
//     Permission is hereby granted, free of charge, to any person obtaining a copy
//     of this software and associated documentation files (the "Software"), to deal
//     in the Software without restriction, including without limitation the rights
//     to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//     copies of the Software, and to permit persons to whom the Software is
//     furnished to do so, subject to the following conditions:
// 
//     The above copyright notice and this permission notice shall be included in all
//     copies or substantial portions of the Software.
// 
//     THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//     IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//     FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//     AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//     LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//     OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//     SOFTWARE

package presentation

import (
	"fmt"
	"io"
	"sort"
)

func (d *displayConfig) displayText(w io.Writer) error {

	sm := d.result.RequiredPermissions

	defaultPerms := sm[d.displayOptions.DefaultResourceGroupResourceID]

	// sort permissions
	// defaultPerms = getUniqueSlice(defaultPerms)
	sort.Strings(defaultPerms)

	// print permissions for default scope
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Println("Permissions Required:")
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------------------")
	for _, perm := range defaultPerms {
		fmt.Println(perm)
	}
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Println()

	if !d.displayOptions.ShowDetailedOutput {
		return nil
	}

	fmt.Println()
	fmt.Println("Break down of permissions by different resource types:")
	fmt.Println()

	// print permissions for other scopes
	for scope, perms := range sm {
		if scope == d.displayOptions.DefaultResourceGroupResourceID {
			continue
		}

		// perms = getUniqueSlice(perms)
		sort.Strings(perms)

		fmt.Printf("Permissions required for %s: \n", scope)
		for _, perm := range perms {
			fmt.Printf("%s\n", perm)
		}
		fmt.Println("--------------")
		fmt.Println()
		fmt.Println()
	}
	return nil
}
