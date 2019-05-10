# microlog

[![Build Status](https://travis-ci.com/lnsp/microlog.svg?token=LqZfkQKrVSBE6pgAPyCG&branch=master)](https://travis-ci.com/lnsp/microlog)

microlog is a dead-simple micrologging engine with an absolute minimalistic user experience. It is usable on mobile, desktop and tablets and aims to be universally usable no matter the handicap.

## Architecture

The application is split up into multiple services. The only one the customer directly accesses is the `gateway` service. It provides a simple server-side rendered web UI for creating and viewing micrologs.

Other services do
- send emails for confirmation and password reset
- creating and verifying user sessions

The goal is to strip down the `gateway` service to a bare minimum, so we can easily replace it later with an REST gateway for a single-page application or similar.