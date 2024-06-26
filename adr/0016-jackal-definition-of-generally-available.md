# 16. Jackal Definition of Generally Available

Date: 2023-05-17

## Status

Accepted

## Context

Eventually, Jackal needs to become a "Generally Available" v1.x.x product that people can rely on for mission-critical operations.  Today, Jackal can be used in these environments; though, it requires someone more involved in the Jackal lifecycle than a normal consumer/user to make that successful due to the regular introduction of breaking changes and the lack of testing in certain areas.

## Decision

To make Jackal a Generally Available product we need to focus on overall stability and mechanisms to ensure long-term stability.  "Stability," in this case, is both that of the features we release and of the APIs and schemas we present to Jackal consumers.

To increase this stability, we decided to implement the following:

- [ ] Mechanism/branching strategy to backport patch fixes to older minor releases
- [x] Clear definition of `released`, `beta`, and `alpha` features, including a matrix of their support across OSes
- [ ] Clear definition of when backward compatibility checks are going to be removed with clear messaging to users
- [ ] End to End testing that covers the `released` features outlined in that feature matrix - this should also be done:
  - across operating systems (specifically: Windows, macOS, Linux)
  - across major k8s distros (specifically: K3d, K3s, Minikube, Kind, EKS, AKS, OpenShift)
  - across registry providers (specifically: Docker Distribution, ECR, ACR)
  - across git providers (specifically: Gitea, GitLab)
- [ ] Unit testing that covers our library code (`src/pkg`) for people using Jackal as a library (code coverage metric TBD)
- [ ] Mechanisms and tests to not break compatibility with packages built with older versions of Jackal
- [ ] Mechanisms to notify users when they may need to upgrade the Jackal Agent (or Pepr capability)
- [ ] Mechanisms to ensure users can easily access documentation specific to the version of Jackal they use
- [ ] Mechanisms to ensure a more seamless Jackal install experience (i.e., macOS binary signing, `sudo apk add jackal`, `asdf install jackal X.X.X`)
- [ ] Regularly published/maintained example package(s) for tutorials/quick install
- [ ] Clear definition/strategy for "what Jackal is" including clear docs on how to use `released` features

>  **Note**: A [x] checkmark denotes items already addressed in Jackal.

## Consequences

Once these are in place, we will have many mechanisms to manage Jackal's stability over time, but we are also signing ourselves up to maintain this promise over the long term, which will increase the burden on the team and reduce our overall velocity - this is good/normal as the project matures, but we will need to recognize that we won't have as much flexibility once we reach GA.

This will also affect how Jackal is supported/marketed beyond the core team, and we should consider how Jackal "GA" will affect those teams and ensure that they are ready to take on any additional burden.
