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

FROM golang:1.18.5@sha256:5540a6a6b3b612c382accc545b3f6702de21e77b15d89ad947116c94b5f42993 as builder

WORKDIR /app
COPY . /app

RUN go get -d -v

# Statically compile our app for use in a distroless container
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app .

FROM gcr.io/distroless/static@sha256:cb0f70353c21d1e472a9ed2055c8a62fd842de52dd2db9dfb31c6e019648bf51

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
