# Proposal for Advanced Proxy Requests
## Table of Contents
* [Summary](#summary)
* [Motivation](#motivation)
    * [Goals](#goals)
* [Proposal](#proposal)
    * <a name="usecases"></a>[Use Cases](#use-cases)
    * <a name="implementationdetails"></a>[Implementation Details](#implementation-details)
    * <a name="risksandmitigations"></a>[Risks and Mitigations](#risks-and-mitigations)
* <a name="graduationcriteria"></a>[Graduation Criteria](#graduation-criteria)
* [Drawbacks](#drawbacks)
* [Alternatives](#alternatives)

## Summary
The current implementation of TXTDirect redirects the user, but does not enable proxy requests. In order to implement proxy request support, including advanced features, it is necessary to first determine whether user configurable proxy requests should be parsed from the configuration or by using txt records. This proposal outlines the pros and cons of using a txt record based approach for enabling proxy requests.

## Motivation
* By moving advanced proxy requests into multiple txt records, it would allow for more features and flexibility such as circumventing the maximum txt record character limit
* Using an instance config goes against the ideas of TXTDirect having data at the customer and using stateless instances

### Goals
* Enable advanced proxy request support without having the data in the instance config
* Support advanced proxying while mitigating potential impacts such as bandwidth, UI and security risks

## Proposal
The proposed solution is to add advanced proxy request support by using multiple txt records that are merged together and parsed within the TXTDirect code.

### Use Cases
* TXTDirect supports all advanced proxy request functionality that is supported by caddy
* By default, the user experience would be unchanged. For more advanced uses, existing users would need to modify the txt records as outlined in the TXTDirect documentation.

### Implementation Details
* Parsing of txt records will be more complex when merging multiple txt records
* The potential for an exposed config (such as one containing authentication) will need to be addressed
* The single txt record character limit would need to be circumvented by using multiple txt records that are merged together
* With this implementation, there are potential bandwidth and UI impacts as well as security risks (whitelisting may be needed)

### Risks and Mitigations
* This section needs more discussion. Whitelisting is an option for mitigating the potential security risks.

## Graduation Criteria
* Verify that all relevant tests pass
* Add additional tests, including e2e tests

## Drawbacks
* More complex txt record parsing will be necessary
* The need to mitigate all the potential performance and security impacts outlined above

## Alternatives
* Enabling advanced proxy requests in the config instance
