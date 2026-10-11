// Copyright 2014 Google Inc.
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

package btree

import (
	"flag"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func init() {
	seed := time.Now().Unix()
	fmt.Println(seed)
	rand.Seed(seed)
}

// perm returns a random permutation of n Int items in the range [0, n).
func perm(n int) (out []Item) {
	for _, v := range rand.Perm(n) {
		out = append(out, Int(v))
	}
	return
}

// rang returns an ordered list of Int items in the range [0, n).
func rang(n int) (out []Item) {
	for i := 0; i < n; i++ {
		out = append(out, Int(i))
	}
	return
}

// all extracts all items from a tree in order as a slice.
func all(t *BTree) (out []Item) {
	t.Ascend(func(a Item) bool {
		out = append(out, a)
		return true
	})
	return
}

// rangerev returns a reversed ordered list of Int items in the range [0, n).
func rangrev(n int) (out []Item) {
	for i := n - 1; i >= 0; i-- {
		out = append(out, Int(i))
	}
	return
}

// allrev extracts all items from a tree in reverse order as a slice.
func allrev(t *BTree) (out []Item) {
	t.Descend(func(a Item) bool {
		out = append(out, a)
		return true
	})
	return
}

var btreeDegree = flag.Int("degree", 32, "B-Tree degree")

func TestBTree(t *testing.T) {
	tr := New(*btreeDegree)
	const treeSize = 10000
	for i := 0; i < 10; i++ {
		if min := tr.Min(); min != nil {
			t.Fatalf("empty min, got %+v", min)
		}
		if max := tr.Max(); max != nil {
			t.Fatalf("empty max, got %+v", max)
		}
		for _, item := range perm(treeSize) {
			if x := tr.ReplaceOrInsert(item); x != nil {
				t.Fatal("insert found item", item)
			}
		}
		for _, item := range perm(treeSize) {
			if !tr.Has(item) {
				t.Fatal("has did not find item", item)
			}
		}
		for _, item := range perm(treeSize) {
			if x := tr.ReplaceOrInsert(item); x == nil {
				t.Fatal("insert didn't find item", item)
			}
		}
		if min, want := tr.Min(), Item(Int(0)); min != want {
			t.Fatalf("min: want %+v, got %+v", want, min)
		}
		if max, want := tr.Max(), Item(Int(treeSize-1)); max != want {
			t.Fatalf("max: want %+v, got %+v", want, max)
		}
		got := all(tr)
		want := rang(treeSize)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
		}

		gotrev := allrev(tr)
		wantrev := rangrev(treeSize)
		if !reflect.DeepEqual(gotrev, wantrev) {
			t.Fatalf("mismatch:\n got: %v\nwant: %v", got, want)
		}

		for _, item := range perm(treeSize) {
			if x := tr.Delete(item); x == nil {
				t.Fatalf("didn't find %v", item)
			}
		}
		if got = all(tr); len(got) > 0 {
			t.Fatalf("some left!: %v", got)
		}
	}
}

func ExampleBTree() {
	tr := New(*btreeDegree)
	for i := Int(0); i < 10; i++ {
		tr.ReplaceOrInsert(i)
	}
	fmt.Println("len:       ", tr.Len())
	fmt.Println("get3:      ", tr.Get(Int(3)))
	fmt.Println("get100:    ", tr.Get(Int(100)))
	fmt.Println("del4:      ", tr.Delete(Int(4)))
	fmt.Println("del100:    ", tr.Delete(Int(100)))
	fmt.Println("replace5:  ", tr.ReplaceOrInsert(Int(5)))
	fmt.Println("replace100:", tr.ReplaceOrInsert(Int(100)))
	fmt.Println("min:       ", tr.Min())
	fmt.Println("delmin:    ", tr.DeleteMin())
	fmt.Println("max:       ", tr.Max())
	fmt.Println("delmax:    ", tr.DeleteMax())
	fmt.Println("len:       ", tr.Len())
	// Output:
	// len:        10
	// get3:       3
	// get100:     <nil>
	// del4:       4
	// del100:     <nil>
	// replace5:   5
	// replace100: <nil>
	// min:        0
	// delmin:     0
	// max:        100
	// delmax:     100
	// len:        8
}

func TestDeleteMin(t *testing.T) {
	tr := New(3)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	for v := tr.DeleteMin(); v != nil; v = tr.DeleteMin() {
		got = append(got, v)
	}
	if want := rang(100); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDeleteMax(t *testing.T) {
	tr := New(3)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	for v := tr.DeleteMax(); v != nil; v = tr.DeleteMax() {
		got = append(got, v)
	}
	// Reverse our list.
	for i := 0; i < len(got)/2; i++ {
		got[i], got[len(got)-i-1] = got[len(got)-i-1], got[i]
	}
	if want := rang(100); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestAscendRange(t *testing.T) {
	tr := New(2)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.AscendRange(Int(40), Int(60), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:60]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendRange(Int(40), Int(60), func(a Item) bool {
		if a.(Int) > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendRange(t *testing.T) {
	tr := New(2)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.DescendRange(Int(60), Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[39:59]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendRange(Int(60), Int(40), func(a Item) bool {
		if a.(Int) < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[39:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendrange:\n got: %v\nwant: %v", got, want)
	}
}
func TestAscendLessThan(t *testing.T) {
	tr := New(*btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.AscendLessThan(Int(60), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rang(100)[:60]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendLessThan(Int(60), func(a Item) bool {
		if a.(Int) > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rang(100)[:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendLessOrEqual(t *testing.T) {
	tr := New(*btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.DescendLessOrEqual(Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[59:]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendlessorequal:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendLessOrEqual(Int(60), func(a Item) bool {
		if a.(Int) < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[39:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendlessorequal:\n got: %v\nwant: %v", got, want)
	}
}
func TestAscendGreaterOrEqual(t *testing.T) {
	tr := New(*btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.AscendGreaterOrEqual(Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.AscendGreaterOrEqual(Int(40), func(a Item) bool {
		if a.(Int) > 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rang(100)[40:51]; !reflect.DeepEqual(got, want) {
		t.Fatalf("ascendrange:\n got: %v\nwant: %v", got, want)
	}
}

func TestDescendGreaterThan(t *testing.T) {
	tr := New(*btreeDegree)
	for _, v := range perm(100) {
		tr.ReplaceOrInsert(v)
	}
	var got []Item
	tr.DescendGreaterThan(Int(40), func(a Item) bool {
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[:59]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendgreaterthan:\n got: %v\nwant: %v", got, want)
	}
	got = got[:0]
	tr.DescendGreaterThan(Int(40), func(a Item) bool {
		if a.(Int) < 50 {
			return false
		}
		got = append(got, a)
		return true
	})
	if want := rangrev(100)[:50]; !reflect.DeepEqual(got, want) {
		t.Fatalf("descendgreaterthan:\n got: %v\nwant: %v", got, want)
	}
}

const benchmarkTreeSize = 10000

func BenchmarkInsert(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkTreeSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		tr := New(*btreeDegree)
		for _, item := range insertP {
			tr.ReplaceOrInsert(item)
			i++
			if i >= b.N {
				return
			}
		}
	}
}

func BenchmarkSeek(b *testing.B) {
	b.StopTimer()
	size := 100000
	insertP := perm(size)
	tr := New(*btreeDegree)
	for _, item := range insertP {
		tr.ReplaceOrInsert(item)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		tr.AscendGreaterOrEqual(Int(i%size), func(i Item) bool { return false })
	}
}

func BenchmarkDeleteInsert(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, item := range insertP {
		tr.ReplaceOrInsert(item)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tr.Delete(insertP[i%benchmarkTreeSize])
		tr.ReplaceOrInsert(insertP[i%benchmarkTreeSize])
	}
}

func BenchmarkDeleteInsertCloneOnce(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, item := range insertP {
		tr.ReplaceOrInsert(item)
	}
	tr = tr.Clone()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tr.Delete(insertP[i%benchmarkTreeSize])
		tr.ReplaceOrInsert(insertP[i%benchmarkTreeSize])
	}
}

func BenchmarkDeleteInsertCloneEachTime(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, item := range insertP {
		tr.ReplaceOrInsert(item)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tr = tr.Clone()
		tr.Delete(insertP[i%benchmarkTreeSize])
		tr.ReplaceOrInsert(insertP[i%benchmarkTreeSize])
	}
}

func BenchmarkDelete(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkTreeSize)
	removeP := perm(benchmarkTreeSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		b.StopTimer()
		tr := New(*btreeDegree)
		for _, v := range insertP {
			tr.ReplaceOrInsert(v)
		}
		b.StartTimer()
		for _, item := range removeP {
			tr.Delete(item)
			i++
			if i >= b.N {
				return
			}
		}
		if tr.Len() > 0 {
			panic(tr.Len())
		}
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkTreeSize)
	removeP := perm(benchmarkTreeSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		b.StopTimer()
		tr := New(*btreeDegree)
		for _, v := range insertP {
			tr.ReplaceOrInsert(v)
		}
		b.StartTimer()
		for _, item := range removeP {
			tr.Get(item)
			i++
			if i >= b.N {
				return
			}
		}
	}
}

func BenchmarkGetCloneEachTime(b *testing.B) {
	b.StopTimer()
	insertP := perm(benchmarkTreeSize)
	removeP := perm(benchmarkTreeSize)
	b.StartTimer()
	i := 0
	for i < b.N {
		b.StopTimer()
		tr := New(*btreeDegree)
		for _, v := range insertP {
			tr.ReplaceOrInsert(v)
		}
		b.StartTimer()
		for _, item := range removeP {
			tr = tr.Clone()
			tr.Get(item)
			i++
			if i >= b.N {
				return
			}
		}
	}
}

type byInts []Item

func (a byInts) Len() int {
	return len(a)
}

func (a byInts) Less(i, j int) bool {
	return a[i].(Int) < a[j].(Int)
}

func (a byInts) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func BenchmarkAscend(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := 0
		tr.Ascend(func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
			}
			j++
			return true
		})
	}
}

func BenchmarkDescend(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := len(arr) - 1
		tr.Descend(func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
			}
			j--
			return true
		})
	}
}
func BenchmarkAscendRange(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := 100
		tr.AscendRange(Int(100), arr[len(arr)-100], func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
			}
			j++
			return true
		})
		if j != len(arr)-100 {
			b.Fatalf("expected: %v, got %v", len(arr)-100, j)
		}
	}
}

func BenchmarkDescendRange(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := len(arr) - 100
		tr.DescendRange(arr[len(arr)-100], Int(100), func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
			}
			j--
			return true
		})
		if j != 100 {
			b.Fatalf("expected: %v, got %v", len(arr)-100, j)
		}
	}
}
func BenchmarkAscendGreaterOrEqual(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := 100
		k := 0
		tr.AscendGreaterOrEqual(Int(100), func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
			}
			j++
			k++
			return true
		})
		if j != len(arr) {
			b.Fatalf("expected: %v, got %v", len(arr), j)
		}
		if k != len(arr)-100 {
			b.Fatalf("expected: %v, got %v", len(arr)-100, k)
		}
	}
}
func BenchmarkDescendLessOrEqual(b *testing.B) {
	arr := perm(benchmarkTreeSize)
	tr := New(*btreeDegree)
	for _, v := range arr {
		tr.ReplaceOrInsert(v)
	}
	sort.Sort(byInts(arr))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := len(arr) - 100
		k := len(arr)
		tr.DescendLessOrEqual(arr[len(arr)-100], func(item Item) bool {
			if item.(Int) != arr[j].(Int) {
				b.Fatalf("mismatch: expected: %v, got %v", arr[j].(Int), item.(Int))
			}
			j--
			k--
			return true
		})
		if j != -1 {
			b.Fatalf("expected: %v, got %v", -1, j)
		}
		if k != 99 {
			b.Fatalf("expected: %v, got %v", 99, k)
		}
	}
}

const cloneTestSize = 10000

func cloneTest(t *testing.T, b *BTree, start int, p []Item, wg *sync.WaitGroup, trees *[]*BTree, lock *sync.Mutex) {
	t.Logf("Starting new clone at %v", start)
	lock.Lock()
	*trees = append(*trees, b)
	lock.Unlock()
	for i := start; i < cloneTestSize; i++ {
		b.ReplaceOrInsert(p[i])
		if i%(cloneTestSize/5) == 0 {
			wg.Add(1)
			go cloneTest(t, b.Clone(), i+1, p, wg, trees, lock)
		}
	}
	wg.Done()
}

func TestCloneConcurrentOperations(t *testing.T) {
	b := New(*btreeDegree)
	trees := []*BTree{}
	p := perm(cloneTestSize)
	var wg sync.WaitGroup
	wg.Add(1)
	go cloneTest(t, b, 0, p, &wg, &trees, &sync.Mutex{})
	wg.Wait()
	want := rang(cloneTestSize)
	t.Logf("Starting equality checks on %d trees", len(trees))
	for i, tree := range trees {
		if !reflect.DeepEqual(want, all(tree)) {
			t.Errorf("tree %v mismatch", i)
		}
	}
	t.Log("Removing half from first half")
	toRemove := rang(cloneTestSize)[cloneTestSize/2:]
	for i := 0; i < len(trees)/2; i++ {
		tree := trees[i]
		wg.Add(1)
		go func() {
			for _, item := range toRemove {
				tree.Delete(item)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	t.Log("Checking all values again")
	for i, tree := range trees {
		var wantpart []Item
		if i < len(trees)/2 {
			wantpart = want[:cloneTestSize/2]
		} else {
			wantpart = want
		}
		if got := all(tree); !reflect.DeepEqual(wantpart, got) {
			t.Errorf("tree %v mismatch, want %v got %v", i, len(want), len(got))
		}
	}
}

func BenchmarkDeleteAndRestore(b *testing.B) {
	items := perm(16392)
	b.ResetTimer()
	b.Run(`CopyBigFreeList`, func(b *testing.B) {
		fl := NewFreeList(16392)
		tr := NewWithFreeList(*btreeDegree, fl)
		for _, v := range items {
			tr.ReplaceOrInsert(v)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dels := make([]Item, 0, tr.Len())
			tr.Ascend(ItemIterator(func(b Item) bool {
				dels = append(dels, b)
				return true
			}))
			for _, del := range dels {
				tr.Delete(del)
			}
			// tr is now empty, we make a new empty copy of it.
			tr = NewWithFreeList(*btreeDegree, fl)
			for _, v := range items {
				tr.ReplaceOrInsert(v)
			}
		}
	})
	b.Run(`Copy`, func(b *testing.B) {
		tr := New(*btreeDegree)
		for _, v := range items {
			tr.ReplaceOrInsert(v)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dels := make([]Item, 0, tr.Len())
			tr.Ascend(ItemIterator(func(b Item) bool {
				dels = append(dels, b)
				return true
			}))
			for _, del := range dels {
				tr.Delete(del)
			}
			// tr is now empty, we make a new empty copy of it.
			tr = New(*btreeDegree)
			for _, v := range items {
				tr.ReplaceOrInsert(v)
			}
		}
	})
	b.Run(`ClearBigFreelist`, func(b *testing.B) {
		fl := NewFreeList(16392)
		tr := NewWithFreeList(*btreeDegree, fl)
		for _, v := range items {
			tr.ReplaceOrInsert(v)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tr.Clear(true)
			for _, v := range items {
				tr.ReplaceOrInsert(v)
			}
		}
	})
	b.Run(`Clear`, func(b *testing.B) {
		tr := New(*btreeDegree)
		for _, v := range items {
			tr.ReplaceOrInsert(v)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tr.Clear(true)
			for _, v := range items {
				tr.ReplaceOrInsert(v)
			}
		}
	})
}

type testRecord struct {
	id    int
	title string
	extra string
}

func testRecordCmp(key int, item testRecord) int {
	if key < item.id {
		return -1
	}
	if key > item.id {
		return 1
	}
	return 0
}

func testRecordPtrCmp(key int, item *testRecord) int {
	if key < item.id {
		return -1
	}
	if key > item.id {
		return 1
	}
	return 0
}

func TestBTreeGDirectKeyLookup(t *testing.T) {
	for _, degree := range []int{2, 3, 4, 32} {
		tr := NewG[testRecord](degree, func(a, b testRecord) bool {
			return a.id < b.id
		})

		const count = 500
		for _, v := range rand.Perm(count) {
			tr.ReplaceOrInsert(testRecord{
				id:    v,
				title: fmt.Sprintf("title-%d", v),
				extra: "payload",
			})
		}

		keyExtractor := func(r testRecord) int { return r.id }

		// Test existing keys across various helper entry points
		for i := 0; i < count; i++ {
			// GetKey with custom comparator
			got, ok := GetKey(tr, i, testRecordCmp)
			if !ok || got.id != i || got.title != fmt.Sprintf("title-%d", i) {
				t.Fatalf("degree %d: GetKey(%d) failed: got (%+v, %v)", degree, i, got, ok)
			}

			// GetKeyOrdered
			gotOrd, okOrd := GetKeyOrdered(tr, i, keyExtractor)
			if !okOrd || gotOrd.id != i {
				t.Fatalf("degree %d: GetKeyOrdered(%d) failed: got (%+v, %v)", degree, i, gotOrd, okOrd)
			}

			// GetKeyBy
			gotBy, okBy := GetKeyBy(tr, i, keyExtractor, Less[int]())
			if !okBy || gotBy.id != i {
				t.Fatalf("degree %d: GetKeyBy(%d) failed: got (%+v, %v)", degree, i, gotBy, okBy)
			}

			// HasKey helpers
			if !HasKey(tr, i, testRecordCmp) {
				t.Fatalf("degree %d: HasKey(%d) returned false", degree, i)
			}
			if !HasKeyOrdered(tr, i, keyExtractor) {
				t.Fatalf("degree %d: HasKeyOrdered(%d) returned false", degree, i)
			}
			if !HasKeyBy(tr, i, keyExtractor, Less[int]()) {
				t.Fatalf("degree %d: HasKeyBy(%d) returned false", degree, i)
			}
		}

		// Test non-existent keys
		for _, missingKey := range []int{-100, -1, count, count + 1, count + 500} {
			if got, ok := GetKey(tr, missingKey, testRecordCmp); ok {
				t.Fatalf("degree %d: expected missing key %d to fail, got (%+v, %v)", degree, missingKey, got, ok)
			}
			if HasKey(tr, missingKey, testRecordCmp) {
				t.Fatalf("degree %d: HasKey(%d) unexpectedly returned true", degree, missingKey)
			}
			if got, ok := GetKeyOrdered(tr, missingKey, keyExtractor); ok {
				t.Fatalf("degree %d: GetKeyOrdered unexpectedly found missing key %d: %+v", degree, missingKey, got)
			}
			if HasKeyOrdered(tr, missingKey, keyExtractor) {
				t.Fatalf("degree %d: HasKeyOrdered unexpectedly returned true for %d", degree, missingKey)
			}
		}
	}
}

func TestBTreeGDirectKeyLookupPointers(t *testing.T) {
	tr := NewG[*testRecord](4, func(a, b *testRecord) bool {
		return a.id < b.id
	})

	const count = 200
	for i := 0; i < count; i++ {
		tr.ReplaceOrInsert(&testRecord{
			id:    i,
			title: fmt.Sprintf("name-%d", i),
		})
	}

	for i := 0; i < count; i++ {
		rec, ok := GetKey(tr, i, testRecordPtrCmp)
		if !ok || rec == nil || rec.id != i {
			t.Fatalf("GetKey pointer lookup failed for id %d", i)
		}
	}

	// Verify zero allocations during direct key search
	allocs := testing.AllocsPerRun(100, func() {
		_, _ = GetKey(tr, 50, testRecordPtrCmp)
	})
	if allocs != 0 {
		t.Fatalf("expected 0 allocations for direct key lookup, got %f", allocs)
	}
}

func TestBTreeGDirectKeyDelete(t *testing.T) {
	for _, degree := range []int{2, 3, 4, 16} {
		tr := NewG[testRecord](degree, func(a, b testRecord) bool {
			return a.id < b.id
		})

		const count = 300
		for _, v := range rand.Perm(count) {
			tr.ReplaceOrInsert(testRecord{id: v, title: fmt.Sprintf("item-%d", v)})
		}

		keyExtractor := func(r testRecord) int { return r.id }

		// Delete non-existent key returns false
		if _, ok := DeleteKey(tr, 9999, testRecordCmp); ok {
			t.Fatalf("degree %d: DeleteKey on non-existent key returned true", degree)
		}
		if tr.Len() != count {
			t.Fatalf("degree %d: length changed after failed delete: %d != %d", degree, tr.Len(), count)
		}

		// Delete elements in random order using distinct delete helpers
		perm := rand.Perm(count)
		for i, k := range perm {
			var deleted testRecord
			var ok bool
			switch i % 3 {
			case 0:
				deleted, ok = DeleteKey(tr, k, testRecordCmp)
			case 1:
				deleted, ok = DeleteKeyOrdered(tr, k, keyExtractor)
			case 2:
				deleted, ok = DeleteKeyBy(tr, k, keyExtractor, Less[int]())
			}

			if !ok || deleted.id != k {
				t.Fatalf("degree %d: failed to delete key %d: got (%+v, %v)", degree, k, deleted, ok)
			}
			if HasKey(tr, k, testRecordCmp) {
				t.Fatalf("degree %d: key %d still present after deletion", degree, k)
			}
			if tr.Len() != count-i-1 {
				t.Fatalf("degree %d: unexpected length %d after deleting %d elements", degree, tr.Len(), i+1)
			}
		}

		if tr.Len() != 0 {
			t.Fatalf("degree %d: expected empty tree after deleting all elements, got len %d", degree, tr.Len())
		}
	}
}

func TestBTreeGDirectKeyIteration(t *testing.T) {
	tr := NewG[testRecord](3, func(a, b testRecord) bool {
		return a.id < b.id
	})

	for i := 0; i < 100; i++ {
		tr.ReplaceOrInsert(testRecord{id: i, title: fmt.Sprintf("val-%d", i)})
	}

	// AscendRangeKey [25, 45)
	var gotAscend []int
	AscendRangeKey(tr, 25, 45, testRecordCmp, func(item testRecord) bool {
		gotAscend = append(gotAscend, item.id)
		return true
	})
	if len(gotAscend) != 20 {
		t.Fatalf("AscendRangeKey count mismatch: %d != 20", len(gotAscend))
	}
	for i, v := range gotAscend {
		if v != 25+i {
			t.Fatalf("AscendRangeKey unexpected element at %d: %d != %d", i, v, 25+i)
		}
	}

	// AscendLessThanKey 15
	var gotLess []int
	AscendLessThanKey(tr, 15, testRecordCmp, func(item testRecord) bool {
		gotLess = append(gotLess, item.id)
		return true
	})
	if len(gotLess) != 15 || gotLess[0] != 0 || gotLess[14] != 14 {
		t.Fatalf("AscendLessThanKey mismatch: %+v", gotLess)
	}

	// AscendGreaterOrEqualKey 85
	var gotGte []int
	AscendGreaterOrEqualKey(tr, 85, testRecordCmp, func(item testRecord) bool {
		gotGte = append(gotGte, item.id)
		return true
	})
	if len(gotGte) != 15 || gotGte[0] != 85 || gotGte[14] != 99 {
		t.Fatalf("AscendGreaterOrEqualKey mismatch: %+v", gotGte)
	}

	// DescendRangeKey [60, 40)
	var gotDescend []int
	DescendRangeKey(tr, 60, 40, testRecordCmp, func(item testRecord) bool {
		gotDescend = append(gotDescend, item.id)
		return true
	})
	if len(gotDescend) != 20 || gotDescend[0] != 60 || gotDescend[19] != 41 {
		t.Fatalf("DescendRangeKey mismatch: len=%d, first=%d, last=%d", len(gotDescend), gotDescend[0], gotDescend[len(gotDescend)-1])
	}

	// DescendLessOrEqualKey 10
	var gotLte []int
	DescendLessOrEqualKey(tr, 10, testRecordCmp, func(item testRecord) bool {
		gotLte = append(gotLte, item.id)
		return true
	})
	if len(gotLte) != 11 || gotLte[0] != 10 || gotLte[10] != 0 {
		t.Fatalf("DescendLessOrEqualKey mismatch: %+v", gotLte)
	}

	// DescendGreaterThanKey 85
	var gotGt []int
	DescendGreaterThanKey(tr, 85, testRecordCmp, func(item testRecord) bool {
		gotGt = append(gotGt, item.id)
		return true
	})
	if len(gotGt) != 14 || gotGt[0] != 99 || gotGt[13] != 86 {
		t.Fatalf("DescendGreaterThanKey mismatch: %+v", gotGt)
	}

	// Early termination test
	var earlyStop []int
	AscendRangeKey(tr, 10, 50, testRecordCmp, func(item testRecord) bool {
		earlyStop = append(earlyStop, item.id)
		return len(earlyStop) < 5
	})
	if len(earlyStop) != 5 {
		t.Fatalf("early stop expected 5 elements, got %d", len(earlyStop))
	}
}

func TestKeyedWrapper(t *testing.T) {
	tr := NewG[testRecord](2, func(a, b testRecord) bool {
		return a.id < b.id
	})

	keyed := NewKeyedOrdered(tr, func(r testRecord) int { return r.id })
	if keyed.Tree() != tr {
		t.Fatalf("Tree() mismatch")
	}

	for i := 0; i < 50; i++ {
		tr.ReplaceOrInsert(testRecord{id: i, title: fmt.Sprintf("k-%d", i)})
	}

	for i := 0; i < 50; i++ {
		item, ok := keyed.Get(i)
		if !ok || item.id != i {
			t.Fatalf("keyed.Get(%d) failed", i)
		}
		if !keyed.Has(i) {
			t.Fatalf("keyed.Has(%d) returned false", i)
		}
	}

	// Test wrapper iteration
	var collected []int
	keyed.AscendRange(10, 20, func(item testRecord) bool {
		collected = append(collected, item.id)
		return true
	})
	if len(collected) != 10 || collected[0] != 10 || collected[9] != 19 {
		t.Fatalf("keyed.AscendRange mismatch: %+v", collected)
	}

	// Test wrapper deletion
	deleted, ok := keyed.Delete(25)
	if !ok || deleted.id != 25 {
		t.Fatalf("keyed.Delete(25) failed")
	}
	if keyed.Has(25) {
		t.Fatalf("keyed.Has(25) returned true after deletion")
	}
}

func TestBTreeGKeyLookupNilAndEmpty(t *testing.T) {
	var nilTree *BTreeG[testRecord]
	if _, ok := GetKey(nilTree, 1, testRecordCmp); ok {
		t.Fatalf("GetKey on nil tree unexpectedly returned true")
	}
	if HasKey(nilTree, 1, testRecordCmp) {
		t.Fatalf("HasKey on nil tree unexpectedly returned true")
	}
	if _, ok := DeleteKey(nilTree, 1, testRecordCmp); ok {
		t.Fatalf("DeleteKey on nil tree unexpectedly returned true")
	}
	AscendRangeKey(nilTree, 0, 10, testRecordCmp, func(item testRecord) bool {
		t.Fatalf("iteration on nil tree called callback")
		return true
	})

	emptyTree := NewG[testRecord](2, func(a, b testRecord) bool { return a.id < b.id })
	if _, ok := GetKey(emptyTree, 1, testRecordCmp); ok {
		t.Fatalf("GetKey on empty tree unexpectedly returned true")
	}
	if HasKey(emptyTree, 1, testRecordCmp) {
		t.Fatalf("HasKey on empty tree unexpectedly returned true")
	}
	if _, ok := DeleteKey(emptyTree, 1, testRecordCmp); ok {
		t.Fatalf("DeleteKey on empty tree unexpectedly returned true")
	}
}

