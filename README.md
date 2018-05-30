# stop optimizing me!

This github repo is sort of a runnable blog post? I promise it gets good.

> Don't run the benchmarks now if you want to avoid spoilers.

### Background story

We changed an API from a format like: 

```json
{
  "p_osx": true,
  "p_linux": false,
  "p_windows": true,
  "p_android": false,
  "can_be_bought": true,
  "has_demo": false,
  "in_press_system": false,
}
```

To a format like:

```json
{
  "traits": ["p_osx", "p_windows", "can_be_bought"]
}
```

... which takes up less space (and is more human-friendly).

But on the client-side, we still need a data representation:

  * that lets us look up quickly whether an object has a particular trait or not
  * that can be easily persisted to a database (where each trait has its own column)

Luckily, Go lets us write custom marshalers/unmarshalers from json, so
we can let it know how we want to decode the `Traits` field of a struct
like this for example:

```go
type Game struct {
  Title string
  Traits GameTraits
}

// create a type alias here, so we can
// define methods on GameTraits
type GameTraits something

func (gt GameTraits) MarshalJSON() ([]byte, error) {
  return nil, errors.New("implement me!")
}

func (gt GameTraits) UnmarshalJSON(data []byte) error {
  return errors.New("implement me too!")
}

// this is a Golang trick to make sure GameTraits implements
// the interfaces we care about
var _ json.Marshaler = GameTraits{}
// we simply declare an unnamed variable of the interface type,
// and assign an empty object of our type to it. If it doesn't
// implement the interface properly, it'll fail at compile time
// (even if our type isn't used anywhere - which isn't that uncommon
// for a library)
var _ json.Unmarshaler = GameTraits{}

```

### Solution 001: Using a map

> See the [001_map_simplest.go](001_map_simplest.go) file for runnable code

A relatively straight-forward solution is to represent our traits as map.

We'll add a type alias for our key type, and add consts for all traits:

```go
type GameTrait string

const (
  GameTraitPlatformWindows GameTrait = "p_windows"
  GameTraitPlatformLinux GameTrait   = "p_linux"
  // etc.
)
```

Then we can define a set of traits as a map with our key type:

```go
// notice the plural here
type GameTraits map[GameTrait]bool
```

> Note: `bool` is overkill here, since we'll only ever have true
> values.
>
> We could use a `map[GameTrait]struct{}` instead, if we wanted
> to actually use this solution - then, values would take no memory at all.

And implement marshalling:

```go
func (gt GameTraits) MarshalJSON() ([]byte, error) {
  // in JSON, our traits are stored in an array of strings,
  // so let's build just that
  var traits []string

  // gt is of type GameTraits, which is a map, so we can iterate
  // over its keys with a for-range:
  for k := range gt {
    // we don't need to check the value stored in the map, since
    // it's always true
    traits = append(traits, k)
  }

  // golang knows how to marshal a string array, so let's let it do
  // the thing. `json.Marshal()` also returns `([]byte, error)`
  // so this works just fine
  return json.Marshal(traits)
}
```

And unmarshalling:

```go
func (gt GameTraits) UnmarshalJSON(data []byte) error {
  // we're given a slice of bytes, let's first unmarshal
  // it as an array of strings, then work from that
  var traits []string

  err := json.Unmarshal(data, &traits)
  if err != nil {
    return err
  }

  // for each element of the array (we don't care about the index)
  for _, trait := range traits {
    // store it in the map
    gt[trait] = true
  }

  // no errors, yay!
  return nil
}
```

Whoops, this version of `UnmarshalJSON` crashes with `assignment to nil map`.

That's because a `map[K]V` is actually a pointer - it need to be created explicitly with `make`.

No problem, let's define `UnmarshalJSON` on a pointer to GameTraits instead,
and make a map from scratch every time it's called.

> Note: Go compilers know to turn `gt.UnmarshalJSON(data)` into `(&gt).UnmarshalJSON(data)`. It's just one of those things you need to know.


```go
// taking a pointer over here
func (gtp *GameTraits) UnmarshalJSON(data []byte) error {
  var traits []string
  // making a fresh map over there
  gt := make(GameTraits)

  err := json.Unmarshal(data, &traits)
  if err != nil {
    return err
  }

  for _, trait := range traits {
    gt[trait] = true
  }

  // make gtp point to our fresh map
  *gtp = gt

  return nil
}
```

And it works pretty well! A very bad and inaccurate microbenchmark shows:

```
5900 ns/op	    1257 B/op	      20 allocs/op
```

### Maps are bad actually

There's a bunch of annoying things about our map implementation.

Let's go through them in no particular order.

The first is **memory layout**. Our `GameTraits` type isn't meant to
be used in the void. It's supposed to be embedded into another type, like so:

```go
type Game {
  Title string
  Traits GameTraits
}
```

In solution 001, `Traits` is a pointer to a map, which means we're actually
dealing with two different objects, at different places in memory,
allocated at different times - and every access means dereferencing the pointer
to the map, hashing the key, finding the bucket, and returning the value.

The second is **ease of use**. Declaring a `GameTraits` map with some
values set isn't the hardest thing in the world thanks to map literals:

```go
traits := &GameTraits{
  // we can use consts as map keys
  GameTraitPlatformLinux: true,
  GameTraitPlatformWindows: true,
}
```

Checking for a value isn't that bad either:

```go
if traits[GameTraitPlatformLinux] {
  // do something linux-specific
}
```

But if we were going to use this approach in the real world, we'd have
defined `GameTraits` like so:

```go
type GameTraits map[GameTrait]struct{}
```

Then declaring and accessing would look like this, respectively:

```go
traits := &GameTraits{
  // uglyyyyyy (but compact in memory)
  GameTraitPlatformLinux: struct{}{},
  GameTraitPlatformWindows: struct{}{},
}

// oh yeah, map lookups actually return (T, bool)
// just another lil' Go thing
if _, ok := traits[GameTraitPlatformLinux]; ok {
  // do something linux-specific
}
```

And, like, ok, this is Go, we're all used to writing more code than we
should be just so the compiler can be fast and the language simple and yadda-yadda, but **come on**.

The third is that **it's not reflect-friendly**. A map is really friendly
to iterate, but that only tells us about the keys currently set - there's
no way to list all possible keys!

> Note: the fact that Go doesn't have *real* enums (or generic types) doesn't
> help our case here at all.

> Note 2: Oh btw, this whole article has an implicit "yes, I've heard about
> language X and it's wonderful, but this is Go we're talking about" agreement.

This is problematic because we're going to end up shoving these values
in an SQLite table, and we definitely want to know what values there are,
as they're each going to have their own column.

### Solution 002: Using a struct

> See the [002_struct_simplest.go](002_struct_simplest.go) file for runnable code

I'm actually sorta proud of this one.

So we start out by defining `GameTraits` as the simplest struct, with
only boolean fields - and we annotate each of them with the string
that appears in the JSON array:

```go
type GameTraits struct {
  PlatformOSX     bool `trait:"p_osx"`
  PlatformWindows bool `trait:"p_windows"`
  // etc.
}
```

Let's review **memory layout** here - remember our type is used in another
type, like so:

```go
type Game {
  Title string
  Traits GameTraits
}
```

This is effectively equivalent to:

```go
type Game {
  Title string
  Traits struct{
    PlatformOSX bool
    PlatformWindows bool
    // etc.
  }
}
```

Which is effectively equivalent to:

```go
type Game {
  Title string
  TraitPlatformOSX bool
  TraitPlatformWindows bool
  // etc.
}
```

Everything gets allocated at the same time, in the same memory block, so
it's cache-friendly and everything.

> Note: ok there's a lot more to consider when we want to design CPU-friendly
> things - not the least of which, alignment. And depending on how often
> traits are accessed, indirection might actually *help* avoid cache misses.
>
> But y'all realize that's way outside the scope of this document right.
> We're doing *microbenchmarks* on a piece of code that we *shouldn't be optimizing*, so let's keep it nice and simple.

Let's do a quick **ease of use** review:

```go
traits := &GameTraits{
  PlatformOSX: true,
  CanBeBought: true,
}

if traits.PlatformLinux {
  // do something linux-specific here
}
```

Beautiful! Well, the beautifulest that Go will allow, but still, that's something.

So, we love it, the CPU loves it (*), all that's left is to implement marshalling
and unmarshalling.

> (*) The CPU might not love it

Now, we don't want to write specific code for each field, because then we'd have
3 places where fields are listed, and it's just too easy to miss one.

Instead, let's use reflection.

> Note: I don't care if you think that it's a bad excuse to use reflection.
> Try and stop me.

It's all relatively straightforward, *provided you've spent weeks getting to
know the specifics of Go reflection*.

```go
func (gt GameTraits) MarshalJSON() ([]byte, error) {
  // eventually we're going to marshal this
  var traits []string

  // let's turn 'gt' into something we can inspect
  val := reflect.ValueOf(gt)
  // let's grab its type too, because we're going to
  // need those `trait:"XXX"` annotations
  typ := val.Type()

  // probably the fastest way to go through all the fields
  // of a struct
  for i := 0; i < typ.NumField(); i++ {
    // this evaluates to true if the ith field of gt is set
    // (in the order in which they were defined - which Go
    // spec says is the same order in which they're laid out in memory)
    if val.Field(i).Bool() {
      // if it's true let's grab the annotation
      trait := typ.Field(i).Tag.Get("trait")

      // ...and add it to our result array
      traits = append(traits, trait)
    }
  }

  // finally we can emit some JSON
  return json.Marshal(traits)
}
```

Unmarshalling is a similar bundle of fun:

```go
// we need to be receiving a pointer, because we're going
// to be modifying the contents of gt
func (gt *GameTraits) UnmarshalJSON(data []byte) error {
  // we have to call `Elem()` here because `gt` is a pointer
  val := reflect.ValueOf(gt).Elem()
  typ := val.Type()

  // let's unmarshal into an array first
  var traits []string

  err := json.Unmarshal(data, &traits)
  if err != nil {
    return err
  }

  for _, trait := range traits {
    // oh no, we have to do an O(n) lookup to find
    // the right field of the struct. I'm sure this won't
    // completely bomb the microbenchmark...
    for i := 0; i < typ.NumField(); i++ {
      // oh noooo, string comparison
      if trait == typ.Field(i).Tag.Get("trait") {
        // at least we're using the fastest way to index fields...
        // small consolation prize
        val.Field(i).SetBool(true)
      }
    }
  }

  // woo we did it
  return nil
}
```

Okay that was a weird value of "fun", but hey, it works.

Let's take a look at the benchmark:

```
 5900 ns/op	    1257 B/op	      20 allocs/op
11700 ns/op	    1392 B/op	      60 allocs/op
```

Oh noooooo. Not only is it 2x slower, it does 3x as many allocations.

> Repeat after me: it *does not matter*. We could ship this and all meet
> at the pub. This is in no way performance-critical. We won't be unmarshalling
> billions or even millions of records in a tight loop. What follows is
> completely gratuitious.
>
> The article should stop right here. But it doesn't.

### Solution 003: bye bye O(n)

> See the [003_struct_cachereflect](003_struct_cachereflect.go) file for runnable code

Okay so, at this point, our beautiful struct-based solution is slower than
a map - and being beaten by a general-purpose hash map is really vexing.

We don't know (because we haven't measured) how expensive exactly reflection is,
but what we **do** know is that for every trait, we have to go through up to
N struct fields and compare strings just to find the right one - and that's,
like, criminal.

> It's not. It's fine. Go home. Stop optimizing me.

But here's the thing: the layout of our `GameTraits` struct does not change.
It's always the same from one execution of `{Unm,M}arshalJSON` to the next.

Yet we always do the same amount of work, as if we didn't know anything about
the type.

I say let's do all the work we can in advance, and cache it in a structure
with an O(1) lookup

> map: So, you've come crawling back...

We'll need a way to map a trait to the index of the correspondingfield in the struct:

```go
// int is fine, we don't have 2^31-ish fields
var gameTraitToIndex map[string]int
```

And we could also use a list of traits (in the same order as the struct),
so we don't need to use reflection to iterate them:

```go
var gameTraits []string
```

We'll use an `init` function (that is guaranteed to run on program startup)
and reflect that struct once and for all:

```go
func init() {
  // making a dummy struct just to get a reflect.Type
  typ := reflect.TypeOf(GameTraitsStruct{})

  // remember maps have to be `make()'d` - assigning to a
  // nil map will crash.
  gameTraitToIndex = make(map[string]int)

  // let's also allocate the gameTraits array instead of
  // using append(), since we know exactly how many items it should contain
  gameTraits = make([]string, typ.NumField())

  for i := 0; i < typ.NumField(); i++ {
    // scanning annotations only once, good!
    trait := typ.Field(i).Tag.Get("trait")
    // all fairly straight-forward bookkeeping here.
    gameTraitToIndex[trait] = i
    gameTraits[i] = trait
  }
}
```

Now, we can make our marshalling function faster - using the `gameTraits`
array to go through each field. We still need to use reflection to see if
it's set, though:

```go
func (gt GameTraitsStruct) MarshalJSON() ([]byte, error) {
  // we're going to marshal that in the end
  var traits []string

  val := reflect.ValueOf(gt)
  // we actually care about the index (i) here, since
  // it's also the index of the field in the GameTraits struct
  for i, trait := range gameTraits {
    // `val.Field(i)` returns a `reflect.Value`, we need to call `Bool()`
    // on it to evaluate it as a boolean.
    if val.Field(i).Bool() {
      // using append is bad (it can cause traits to be reallocated several
      // times if we're unlucky), but apart from iterating gameTraits twice,
      // we can't really do much better
      traits = append(traits, trait)
    }
  }
  return json.Marshal(traits)
}
```

Similarly, we can write a better unmarshaler:

```go
func (gt *GameTraitsStruct) UnmarshalJSON(data []byte) error {
  // ok this part is second nature by now
  var traits []string
  err := json.Unmarshal(data, &traits)
  if err != nil {
    return err
  }

  // we have a pointer receiver (the (gt *GameTraitsStruct) bit in our
  // function declaration), so we need to call `Elem()` here
  val := reflect.ValueOf(gt).Elem()
  for _, trait := range traits {
    // our handy map lets us know which field to set in O(1)
    // looks pretty good!
    val.Field(gameTraitToIndex[trait]).SetBool(true)
  }
  return nil
}
```

Let's check the benchmarks:

```
 5900 ns/op	    1257 B/op	      20 allocs/op
11700 ns/op	    1392 B/op	      60 allocs/op
 5300 ns/op	    1072 B/op	      20 allocs/op
```

Wow, this is much better! "Ship it" levels of better.

It even beats the map approach by a hair - with the same number of
memory allocations (20), and slightly lower memory usage.

For real this time, **this is where the article should stop**. We took
a dumb approach, it was slower than another dumb approach, so we made
it slightly smarter *without making an unholy mess*, and now it consistently
beats the first dumb approach.

There **is** such a thing as good enough, and that is most definitely it
right here. Please, please stop reading here.

### Solution 004: it's all bytes in the end

Oh.. you're still here.

> See the [004_struct_handrolled](004_struct_handrolled.go) file for runnable code

So 20 allocations to marshal and unmarshal a bunch of traits feels like it's too much.

> It's not. It's not too much. Turn back, it's still time!

...after all, we're building a whole `[]string` just to call some function
of the standard Go library and to throw it away.

That's bad!

> It's not that bad.

No, it is! We're not even pooling it, so it's a fresh allocation every time.
That means we're generating garbage the GC will have to free eventually - it'll
have to keep track of these temporary objects, and reclaim then in a sweep phase,
and..

> THAT'S THE GC'S JOB. IT'S FIIINE.

We just need to return a `[]byte`, right? Why don't we build that directly?

> Because...

And JSON is not that hard

> ...it really is though.

Okay, JSON done right is really tricky, but we don't care about all valid JSON,
we only care about:

  * an array
  * of values that can only contain `[A-Za-z0-9_]`

I'm sure we can parse and emit that easily!

> Ok buddy you're on your own

Let's goooooooooooo

```go
func (gt GameTraitsStruct) MarshalJSON() ([]byte, error) {
  // ok, this is the thing we *won't* handroll: bytes.Buffer
  // is pretty well-optimized, it'll be hard to beat.
  // we even allocate it on the stack, and it has a `bootstrap [64]byte`
  // field that will *probably* be enough for most calls
  var bb bytes.Buffer

  // starting a JSON array, dum-de-dum
  bb.WriteByte('[')

  first := true
  val := reflect.ValueOf(gt)
  for i, trait := range gameTraits {
    // still 
    if val.Field(i).Bool() {
      if first {
        // if it's the first value we're writing,
        // the next value will not be the first anymore
        first = false
      } else {
        // if it's *not* the first value we're writing,
        // we need a comma separator
        bb.WriteByte(',')
      }
      // JSON strings are double-quoted
      bb.WriteByte('"')
      // we don't need to escape anything, as long as our
      // values are [A-Za-z0-9_]
      bb.WriteString(trait)
      // let's not forget to close the string
      bb.WriteByte('"')
    }
  }
  
  // ...and close the array
  bb.WriteByte(']')

  // oh btw we completely did forgo error handling
  // we're writing to memory, what could wrong?
  // (ok, we could be out of RAM, but then we'd
  // have bigger problems)

  // tada! no json.Marshal()!
  return bb.Bytes(), nil
}
```

Ok that wasn't so bad.

What I mean is that the unmarshaller is much worse still:

```go
func (gt *GameTraitsStruct) UnmarshalJSON(data []byte) error {
  val := reflect.ValueOf(gt).Elem()
  // oh no that's never a good sign
  i := 0
  // ok so I guess our function will accept incomplete inputs
  // that's fine (:fire:)
  for i < len(data) {
    switch data[i] {
    case '"':
      // oh look a string started, let's find the matching double quote
      j := i + 1
    scanString:
      for {
        switch data[j] {
        case '"':
          // we found the matching double quote!
          // better hope we don't have an off-by-one error here
          // (there isn't, but there was the first time I wrote this)
          trait := string(data[i+1 : j])
          // i is our main cursor, skip over the whole quoted string
          i = j + 1
          // still using our map of "trait name to field index" to
          // speed things up.
          // still using reflection though.
          val.Field(gameTraitToIndex[trait]).SetBool(true)
          // oh yeah go has labels, very handy
          // probably unneeded here, but years of C/JS 
          // have made me paranoid about switch statements.
          break scanString
        default:
          // if we have anything other than a double quote, keep reading.
          // this would break *badly* if double quotes were allowed
          // in our trait names, because we do not handle escaping at all.

          // in fact, our input could be valid but made entirely of `\uXXXX`
          // escapes and this wouldn't handle it correctly.

          // it's unlikely in the real world, but let's just say we're not
          // writing a JSON-compliant parser - we're writing for a very
          // narrow subset.
          j++
        }
      }
    case ']':
      // ah, that must mean the array is terminated (fingers crossed)
      // so many checks we're not making here, I can't even...

      // oh btw, notice that we never checked if the array *started*...
      // in other terms, our function will be happy with this input:
      // 
      //   "a",DINOSAUR"b"]haha whoops trailing data
      //
      // which is fantastic and disgusting. it's fantasting.
      return nil
    default:
      i++
    }
  }
  
  // is it lunch time already
  return nil
}
```

At this point in the game, we better find a good lawyer, because I've
lost track of how many crimes we've committed.

But let's run benchmarks:

```
 5900 ns/op	    1257 B/op	      20 allocs/op
11700 ns/op	    1392 B/op	      60 allocs/op
 5300 ns/op	    1072 B/op	      20 allocs/op
  700 ns/op	     128 B/op	       3 allocs/op 
```

**B e a u t i f u l**.

We're doing **a sixth** of the allocations, using **a tenth** of the memory,
and are running **almost 10x faster**.

Surely we can stop there right.

> YOU COULD HAVE STOPPED EIGHT PAGES AGO

### Solution 005: oh god why

...ok let's try one more thing.

> See the [005_unreasonably_custom](005_unreasonably_custom.go) file for runnable code

I'm sorta bothered by solution 004, because it's dirty and bad, but it's not
100% dirty and bad. We're still using reflection, which is one of the Great Evils
and the source of all misery on earth and beyond.

Surely we can go, like, much further into stupidity.

I know I talked about code duplication before, and how we shouldn't be listing
traits in three different places but hey fuck it let's do exactly that.

```go
func (gt GameTraitsStruct) MarshalJSON() ([]byte, error) {
  // we know that part
  var bb bytes.Buffer
  bb.WriteByte('[')

  first := true
  // let's have one of these blocks for each value, ok sure why not
  if gt.PlatformAndroid {
    if first {
      first = false
    } else {
      bb.WriteByte(',')
    }
    // oh neat, we turned 3 calls into one!
    bb.WriteString(`"p_android"`)

    // ...we could even have `,"p_android"` in a branch
    // ...no let's just finish this freakin' post.
  }
  // yeah one more
  if gt.PlatformWindows {
    if first {
      first = false
    } else {
      bb.WriteByte(',')
    }
    bb.WriteString(`"p_windows"`)
  }
  // and so on and so forth

  // <snip.......
  //
  // SO MUCH CODE OMITTED
  //
  // ........snip>

  // aw yiss bb
  bb.WriteByte(']')
  return bb.Bytes(), nil
}
```

That's just marshalling though!

Surely we can do something equally stupid for unmarshalling?

Something along the lines of

```go
switch trait {
  case "p_osx":     gt.PlatformOSX = true
  case "p_windows": gt.PlatformWindows = true
  // etc.
}
```

No. No no no.

See, a sufficiently smart compiler would generate efficient code for that
(instead of doing up to N string comparisons).

But we're writing Go here, so let's not assume the compiler is sufficiently
smart.

Instead, let's write exactly what we would expect a smart compiler to generate:

```go
func (gt *GameTraitsStruct) UnmarshalJSON_UnreasonablyCustom(data []byte) error {
  // same boring parser as solution 004, skip until next comment, right...
  i := 0
  for i < len(data) {
    switch data[i] {
    case '"':
      j := i + 1
    scanString:
      for {
        switch data[j] {
        case '"':
          // ...there. welcome back!
          trait := data[i+1 : j]
          // `trait` is now a byte slice ([]byte) that holds, well, our trait.
          // let's not compare strings, that's dumb.
          // let's compare as little as possible
          switch trait[0] {
          case 'p':
            // if it starts with a `p`, it's a platform.
            // we know trait[1] is always going to be '_', so
            // let's not bother checking it
            switch trait[2] {
            case 'w':
              // luckily all platforms start with a unique letter!
              gt.PlatformWindows = true
            case 'l':
              gt.PlatformLinux = true
            case 'o':
              gt.PlatformOSX = true
            case 'a':
              gt.PlatformAndroid = true
            }
          // only "has_demo" starts with 'h'
          case 'h':
            gt.HasDemo = true
          // etc.
          case 'c':
            gt.CanBeBought = true
          case 'i':
            gt.InPressSystem = true
          }
          // rest of the boring parser omitted for brevity (lol)
}
```

I mean, if *I* was a compiler, that's what I would generate. Maybe the ordering
would be different, but maybe the CPU's pipelining is good enough that it doesn't
matter. I would do some profile-guided optimization (PGO) if the pay was good
enough, but I hear most compilers are free these days, so I'm not holding my digital breath.

Does this perform as well as it looks ugly?

```
 5900 ns/op	    1257 B/op	      20 allocs/op
11700 ns/op	    1392 B/op	      60 allocs/op
 5300 ns/op	    1072 B/op	      20 allocs/op
  700 ns/op	     128 B/op	       3 allocs/op 
  300 ns/op	     112 B/op	       1 allocs/op
```

OH YOU BET IT DOES.

---

Ok so this is a good example of what *not* to do.

I was curious how fast I could get it, and I had already written that code
(which I'm now going to trash), I figured I might as well turn it into an
exploratory lesson of what *not* to do.

Take care y'all, I'm off shipping solution 003 because it's better than good enough.

 * You can [follow me on Twitter](https://twitter.com/fasterthanlime) if you like this

(Please do, so this won't have been all in vain)

> Note: to run the benchmarks for yourself, clone this repo
> and run `go test -benchmem -bench .`
>
> In the paragraphs above, I've rounded the `ns/op` values liberally
> to make comparison easier
>
> These microbenchmarks are never reliable to start with, so I hope you'll forgive me.
