# [J]SON [A]PI [T]EST

This module contains some util functions for unit testing JSON API handler, see [Features](#features).

Only use in unit test, especially for JSON API testing.
This module not guarantee for others API.

## Getting Started

### Features

- Features related to **http.Request**:
    - NewRequest with JSON body
    - Add JSON Body
    - Add Path Params with URL template
    - Add Header
    - Add Query
    - build *http.Request with fluent interface

- Features related to **httptest.ResponseRecorder**
    - Assert status
    - Assert JSON body

### Usage example

[//]: <> (### Prerequisites)

[//]: <> (### Installing)

## Running the tests

```bash
go test -cover
```

[//]: <> (### Break down into end to end tests)

[//]: <> (### And coding style tests)

[//]: <> (## Built With)

[//]: <> (## Contributing)

[//]: <> (## Versioning)

## Authors

* **VictorNM** - *Initial work* - [VictorNM](https://github.com/VictorNM)

[//]: <> (## License)

[//]: <> (## Acknowledgments)
