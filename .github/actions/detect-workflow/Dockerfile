# Copyright 2022 SLSA Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.19.2@sha256:b850621230956a6d960d6d7cfaba6a8a2e8e245b230a928ef66aa0cfd065e229 as builder

WORKDIR /app
COPY . /app

RUN go get -d -v

# Statically compile our app for use in a distroless container
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app .

FROM gcr.io/distroless/static@sha256:f4787e810dbc39dd59fcee319cf88e8a01181e1758dbd07c32ed4e14a9ba8904

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
