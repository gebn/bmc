# Baseboard Management Controller

[![Build Status](https://travis-ci.org/gebn/bmc.svg?branch=master)](https://travis-ci.org/gebn/bmc)
[![GoDoc](https://godoc.org/github.com/gebn/bmc?status.svg)](https://godoc.org/github.com/gebn/bmc)
[![Go Report Card](https://goreportcard.com/badge/github.com/gebn/bmc)](https://goreportcard.com/report/github.com/gebn/bmc)

This project implements an IPMI v1.5/2.0 remote console in pure Go, to interact with BMCs.

## Specifications

All references to sections in code use the following documents:

 - ASF
    - [v2.0](https://www.dmtf.org/sites/default/files/standards/documents/DSP0136.pdf)
 - DCMI
    - [v1.0](https://www.intel.com/content/dam/www/public/us/en/documents/technical-specifications/dcmi-spec.pdf)
    - [v1.1](https://www.intel.com/content/dam/www/public/us/en/documents/technical-specifications/dcmi-v1-1-rev1-0-spec.pdf)
    - [v1.5](https://www.intel.com/content/dam/www/public/us/en/documents/technical-specifications/dcmi-v1-5-rev-spec.pdf)
 - IPMI
    - [v1.5](https://www.intel.com/content/dam/www/public/us/en/documents/product-briefs/second-gen-interface-spec-v1.5-rev1.1.pdf)
    - [v2.0](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.pdf)

## Contributing

Contributions in the form of bug reports and PRs are greatly appreciated.
Please see [`CONTRIBUTING.md`](CONTRIBUTING.md) for a few guidelines.
