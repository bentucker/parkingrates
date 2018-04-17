# Parking Rates

This application calculates parking rates for a given time period,
based on a pre-defined fee schedule. Overnight parking is not currently
supported.

## Dependencies

* Protobuf 3.5.1
* gRPC 1.11.0
* Go 1.10.1

## Building

The above dependencies must all be installed and the binaries must be
in the $PATH.

This project uses dep for dependency management. You can install dep
and download all project dependencies by running

> make deps

Once dep has finished fetching all the dependencies, you can build the
application by running

> make

## Usage
Run the parking rates service
> ./parkingrates

By default, this will start the gRPC service listening on port 32884
and the REST service listening on port 32885.

For more advanced options, see `parkingrates --help`.

Now that the parking rates service is running, you can query it via curl

> curl -X POST -H "Content-Type: application/json" http://localhost:32885/v1/spothero/getrates -d '{"start": "2015-07-04T07:00:00Z", "end": "2015-07-04T20:00:00Z"}'