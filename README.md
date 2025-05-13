# Maceo - Analyze and anonymize data in real-time

Maceo is an [OpenFaaS](https://www.openfaas.com/) serverless function designed
to **analyze** and **anonymize** personally identifiable information (PII) in
real-time. It leverages the powerful capabilities of
[Microsoft Presidio](https://github.com/microsoft/presidio) to detect and
redact sensitive data such as names, phone numbers, emails, and more.

## Features

- 🧠 **PII Detection** – Uses Presidio Analyzer to identify PII in text
- 🛡️ **Anonymization** – Applies Presidio Anonymizer to redact detected PII
- ⚡ **Real-Time Functionality** – Runs as an OpenFaaS function for scalable, on-demand usage
- 🔧 **Configurable** – Customize what types of PII to detect

## Use Cases

- Data privacy compliance (e.g., GDPR, HIPAA)
- Real-time chat or support log redaction
- Secure logging or auditing pipelines
- Anonymizing customer feedback or user-generated content

## Architecture

Maceo serves as a wrapper function that accepts a plain text, sends the input
to Presidio for analysis and anonymization and returns the redacted version of
the original text.

## Getting Started

### Prerequisites

- [OpenFaaS](https://docs.openfaas.com/deployment/) installed and running
- [Microsoft Presidio - For PII anonymization in text](https://microsoft.github.io/presidio/installation/#using-docker) installed and running

### Deploying with OpenFaaS

```bash
export OPENFAAS_URL=https://openfaas.example.com
faas-cli secret create --from-literal='{}' maceo
faas-cli deploy --image=ghcr.io/tschaefer/maceo:0.1.0 --name=maceo
```
This creates the required empty configuration secret for the function and
deploys the function to your OpenFaaS instance.

See [Blog Post](https://blog.tschaefer.org) for a detailed guide on how to set
up with [OpenFaaS Edge](https://docs.openfaas.com/deployment/edge/).

## Usage

Send a POST request to the function with a plain text payload.

```bash
curl --include --request POST https://maceo.example.com \
  --header "Content-Type: text/plain; charset=utf-8" \
  --data-binary  "@contrib/sample"
HTTP/2 200
content-type: text/plain; charset=utf-8
date: Tue, 13 May 2025 05:47:25 GMT
x-call-id: fa2bf983-869a-4bca-b68b-cfb7a3140d0c
x-duration-seconds: 0.078652
x-maceo-commit: ca7eea42494e47927860f2493ef05f54fbfd3a1f
x-maceo-version: 0.1.0
x-openfaas-eula: openfaas-ce
x-served-by: openfaas-ce/0.27.12
x-start-time: 1747115245869918778
content-length: 1157

Sample Job Application Letter
Ms. <PERSON>
DSC Company
68 <LOCATION>, CA 09045
<PHONE_NUMBER>

May 11, 2025

Dear Ms. <PERSON>,

I am writing this letter to apply for a junior programmer position advertised in your organisation. As requested, I am enclosing a completed job application, my certificates, my resumes, and four references in this letter.

The opportunity presented in this listing is exciting. I believe that my firm and years of technical experiences and education will make me a competent person for the position. The main strengths that I have, which I will contribute to this position include:

I have designed, developed and supported many different live use applications.
I continuously work towards achieving my goals through hard work and excellence.
I provide exceptional contributions to the needs and wants of the consumers.
I have a Bachelor of Science degree in Computer Programming. Additionally, I have in-depth knowledge of the complete cycle of a soft development project. Whenever the need arises, I learn new technologies.
I can be reached on <PHONE_NUMBER>.
Thank you for your time and consideration.

Sincerely,

<PERSON>
```

## Configuration

The function can be configured using a secret. The secret is a JSON object
containing the following keys:

- `upstreams`: The URLs of the Presidio Analyzer and Anonymizer services.
    - `analyze`: The URL for the Presidio Analyzer service. (default: `http://10.62.0.1:5001`)
    - `anonymize`: The URL for the Presidio Anonymizer service. (default: `http://10.62.0.1:5002`)
- `entities`: The list of PII entities to detect and anonymize. (default: `nil`)
- `language`: The language to use for PII detection. (default: `en`)
- `score_threshold`: The minimum score for a PII entity to be considered detected. (default: `0.0`)

Find an example configuration in the `contrib/config` file.

## Health Check

The function exposes a health check endpoint at `/health`. The check verifies
the syntax of the configuration secret and the availability of the Presidio
services. It returns a 200 OK status if everything is working correctly and a
500 status if there are any issues.

```bash
curl --include --request GET https://maceo.example.com/health
HTTP/2 200
content-type: text/plain
date: Tue, 13 May 2025 06:43:24 GMT
x-call-id: 0985e928-fc7d-454a-aab2-620c4c32e533
x-duration-seconds: 0.007149
x-maceo-commit: ca7eea42494e47927860f2493ef05f54fbfd3a1f
x-maceo-version: 0.1.0
x-openfaas-eula: openfaas-ce
x-served-by: openfaas-ce/0.27.12
x-start-time: 1747118604460379210
content-length: 0
```

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.
For major changes, open an issue first to discuss what you would like to change.

Ensure that your code adheres to the existing style and includes appropriate tests.

## License

This project is licensed under the [MIT License](LICENSE).
