# How you should write unit tests in GoLang.

Sometimes, I have to interview candidates for software developer roles in Python & GoLang. One of the hardest part to
get right, is to evaluate the seniority level of the candidate.

I can determine it by asking different questions, like:

* What is the difference between git merge & git rebase?
* How do you optimize the size of a Dockerfile?

But one of my favourite thing to ask is the difference between a unit test, and an integration test, or even how much
time should a unit test take to run.

A lot of inexperienced candidates think that they are writing unit tests, when they are actually integration ones, since
they often require:

* A database, with tear up / tear down (usually SQLite).
* Real http requests.
* A web server.
* etc ...

A unit test should be standalone.

It generally does not require anything else than code in the language you're writing.

One of the key element which should give a hint that you're doing something wrong, is the time these tests take to run.

Things like DB Connection or HTTP Requests take time, and that's another reason why they should not be in the
picture.

Unit tests should be :

* Fast, to give an immediate feedback.
* Simple, and then easy to fix when they break.
* Readable, and then useful as a documentation.
* Stateless, which means that a test should always produce the same result.

Usually, when you understand that, the code is already made, and impossible to test: it's too late. Which brings one of
the biggest requirements: your code has to be designed to support unit tests, and that's one of the reason why Test
Driven Development is popular.

In languages like Python, that's easy. You can override anything: there is usually a dirty trick to reach your goal.

But in GoLang, that's not possible. So how do we do that ?

For unit tests in GoLang, you'll be happy to use two things:

* Interfaces
* Mocks

I highly recommend using these two libraries:

* [Testify](https://github.com/stretchr/testify). This lib gives useful building blocks to create your tests like
  assertions, test suites, etc.
* [GoMock](https://github.com/golang/mock). GoMock is both a library, and a client to generate Mocks, based on
  interfaces.

Now let's see an example. 

Let's say I want to create a little command line giving me the ability to get my IP
using [ipify](https://www.ipify.org/), and write it into a file.

I could go straight away, and write something like this:

```golang
package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func Execute(outputFile string) error {
	//Get request on the API.
	resp, err := http.Get("https://api.ipify.org?format=json")

	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	//Decode the JSON.
	var result map[string]string

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	//Write its content to a file.
	return ioutil.WriteFile(outputFile, []byte(result["ip"]), 0644)
}

func main() {
	if err := Execute(os.Args[1]); err != nil {
		log.Fatal(err)
	}
}
```

And then it works:

```bash
> go run main.go output.txt && cat output.txt
184.161.4.105
```

Now, how do we test this `Execute` function ?

We can identify two annoying lines :

```golang
resp, err := http.Get("https://api.ipify.org?format=json")
```

and

```golang
ioutil.WriteFile(outputFile, []byte(result["ip"]), 0644)
```

* What does happen to `http.Get` if the API is down ?
* What if I want to test an error ?
* What about `WriteFile` if I have several tests running ? I don't want to see files popping everywhere on my machine.

That's when you need to reorganize your code, and decouple these implementations from the routines consuming them.

For that, we will declare two interfaces, that we keep along with the `Execute` function.

```golang
package main

type IPGetter interface {
	GetPublicIP() (*string, error)
}

type FileWriter interface {
	Write(filename string, data []byte, perm fs.FileMode) error
}
```

Always put the interfaces where they are consumed. It allows you to only declare what you need (here `GetPublicIP`
and `Write`).

The implementations can continue to evolve: these interfaces ensure that they keep implementing **at least** these two methods.

Now, the Execute function can consume these interfaces directly, and does not know any more about their implementation 
(only the interface as an **abstraction**).

```golang
package main

type IPGetter interface {
	GetPublicIP() (*string, error)
}

type FileWriter interface {
	Write(filename string, data []byte, perm fs.FileMode) error
}

func Execute(
	ipGetter IPGetter,
	fileWriter FileWriter,
	outputFile string,
) error {
	//Get request on the API.
	publicIP, err := ipGetter.GetPublicIP()

	if err != nil {
		return err
	}

	//Write its content to a file.
	return fileWriter.Write(outputFile, []byte(*publicIP), 0644)
}
```

We can now create the two structs implementing them (we'll put them in an `internal` package).

`internal/file.go`:
```golang
package internal

type FileRepository struct{}

func (f *FileRepository) Write(filename string, data []byte, perm fs.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}
```
`internal/ipify.go`:
```golang
package internal

type IpifyService struct{}

func (i *Ipify) GetPublicIP() (*string, error) {
	resp, err := http.Get("https://api.ipify.org?format=json")

	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	//Decode the JSON.
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	ip := result["ip"]
	return &ip, nil
}
```

And now, the technique is to inject these implementations in our `Execute` function, since they satisfy the required
interfaces.

```golang
package main

func Execute(
	ipGetter IPGetter, fileWriter FileWriter, outputFile string,

) error {
	// [...]
}

func main() {
	if err := Execute(&internal.Ipify{}, &internal.FileRepository{}, os.Args[1]); err != nil {
		log.Fatal(err)
	}
}
```

The beauty of this technique, is that I can switch the implementation of `IPGetter` and `FileWriter` for anything I
want.

In a real life scenario, we pass the real implementation, but for unit tests, we'll be able to pass Mocks.

To sum up, currently we have two packages.

* main (at the root)
    * Contains the Execute function and the interfaces.
* internal (sub folder)
    * Contains the two implementations to interact with the file system and ipify.

```bash
.
├── go.mod
├── go.sum
├── internal
│   ├── file.go
│   └── ipify.go
├── main.go
└── README.md
```

Let's run this code :

```bash
> go run main.go output.txt && cat output.txt
184.161.4.105
```

Still working !

Now finally, we can write our tests !

We create a `main_test.go` file.

```golang
package main

import "testing"

func TestExecute(t *testing.T) {
	// err := Execute(?, ?, "output.txt")
}
```

What now ? We need some implementation for these interfaces allowing us to test the function without doing the real
deal.

That's where [GoMock](https://github.com/golang/mock) is great.

You can generate a Mock implementation, by providing the interfaces defined in `main.go`.

```bash
mockgen -source main.go \
  -package=main \
  -destination main_mock.go \
  IPGetter,FileWriter
```

It will generate a file `main_mock.go`, with two mocks implementations.

You can now simply invoke them from our tests :

```golang
package main

func TestExecute(t *testing.T) {
	// A Controller represents the top-level control of a mock ecosystem.
	ctrl := gomock.NewController(t)
	// Create the mocks.
	mockIpGetter := NewMockIPGetter(ctrl)
}
```

You can program them to receive a certain set of parameters, and return an associated result.
GoMock can even validate how many times the method is called.

```golang
package main

func TestExecute(t *testing.T) {
	expectedIP := "184.162.7.66"
	// [...]
	// I expect GetPublicIp to return the IP above.
	mockIpGetter.EXPECT().
		GetPublicIP().
		Return(&expectedIP, nil).
		Times(1)
}
```

The only thing left to do is to pass the mocks implementations to your `Execute` function, the exact same
way we did with the real implementations in the `main` function.

```golang
package main

func TestExecute(t *testing.T) {
	// [...]
	err := Execute(mockIpGetter, mockFileWriter, expectedOutputFile)
}
```

Here the full example :

```golang
package main

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"testing"
)

func TestExecute(t *testing.T) {
	// A Controller represents the top-level control of a mock ecosystem.
	ctrl := gomock.NewController(t)
	// Create the mocks.
	mockIpGetter := NewMockIPGetter(ctrl)
	mockFileWriter := NewMockFileWriter(ctrl)

	expectedOutputFile := "output.txt"
	expectedIP := "184.162.7.66"

	// I expect GetPublicIp to return the IP above.
	mockIpGetter.EXPECT().
		GetPublicIP().
		Return(&expectedIP, nil).
		Times(1)

	// I expect this ip to be written in the file output.txt.
	mockFileWriter.EXPECT().
		Write(expectedOutputFile, []byte(expectedIP), fs.FileMode(0644)).
		Return(nil).
		Times(1)

	// Run the code.
	err := Execute(mockIpGetter, mockFileWriter, expectedOutputFile)

	// Ensure there are no errors for this scenario.
	assert.NoError(t, err, "no errors expected")

	//Ensure that all the expected mocks have been called.
	ctrl.Finish()
}
```

Run the tests using :

```bash
go test ./...
```

You can also get the coverage by using :

```bash
go test ./... --coverprofile cover.out
go tool cover -html=cover.out
```

The only thing left to do is to get the maximum coverage !

Of course this is a simple example, but imagine using this technique for an `Execute` function with a lot of logic and
edge cases. That's a game changer.

You can also test directly your HTTP / DB classes and others, using specialized mock packages like
[go-sqlmock](https://github.com/DATA-DOG/go-sqlmock). 

But usually, you want to keep your implementations to the strict minimum, 
like a thin wrapper: then unit tests for this core part are still hard to write, but do not bring that much value. 

Finally, this last gap can be covered by integration tests, but that's for another story.

I hope this helps. Good luck.
