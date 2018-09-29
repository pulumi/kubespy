# `kubespy` roadmap

The mission of `kubespy` is to provide tools for understanding what is happening to Kubernetes
resources in real time. This document contains ideas for what we think that _should_ involve, but if
you see something missing we'd love to hear about it, and we encourage you to [open an
issue][new-issue] to tell us about it!

## Make `kubespy` "a real project"

`kubespy` began life as a tiny tool to make gifs that show how Kubernetes works. We did not expect
users to find this interesting enough to use.

But, now they do, and the time has come to turn this project from a one-off, to something with
engineering processes more befitting the trust users place in it:

-   Standard contrib docs (`CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `ROADMAP.md`)
-   CI integration
-   Tests

## Make the tool usable to non-experts

We were surprised when people wanted to use `kubespy`. We were even more surprised when the people
who wanted to use `kubespy` were confused by things that we (foolishly!) took for granted. Stuff
like: users don't know what `dep` is, or why `v1` must go in front of `Pod`, or whether the
kubeconfig file can have multiple clusters. In short, users who have to use Kubernetes, but who are
not experts.

To make this tool effective for this audience it is critical that we make it simple to install and
use:

-   Install must be turn-key on all common platforms. `brew`, `apt`, `chocolatey`, and binary
    releases are a must.
-   CLI experience must take a minimal amount of context to produce sensible output.
    `kubespy status po` should work, as should `kubespy status v1 Pod`.
-   Errors should prescribe action where appropriate.

## Expand `kubespy trace` to include all commonly-used "complex" types

`kubespy trace` is meant to display a real-time, human-readable, high-level picture of the status of
"complex" Kubernetes resource types. For example, the status of a `Service` is distributed across
the `Service` object itself, an `Endpoints` object of the same name that contains information about
how to direct traffic to `Pod`s, and the `Pod`s themselves.

Currently `kubespy trace` supports _only_ `Service`. But also of interest to our users are:

-   `Deployment`
-   `ReplicaSet`
-   `Ingress`
-   `StatefulSet`

[new-issue]: https://github.com/pulumi/kubespy/issues/new
