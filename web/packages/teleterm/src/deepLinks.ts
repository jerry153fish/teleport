/**
 * Copyright 2023 Gravitational, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
import * as whatwg from 'whatwg-url';

export const CUSTOM_PROTOCOL = 'teleport' as const;

export type DeepLinkParseResult =
  // Just having a field like `ok: true` for success and `status: 'error'` for errors would be much more
  // ergonomic. Unfortunately, `if (!result.ok)` doesn't narrow down the type properly with
  // strictNullChecks off. https://github.com/microsoft/TypeScript/issues/10564
  | { status: 'success'; url: DeepURL }
  | ParseError<'malformed-url', { error: TypeError }>
  | ParseError<'unknown-protocol', { protocol: string }>
  | ParseError<'unsupported-uri'>;

type ParseError<Reason, AdditionalData = void> = AdditionalData extends void
  ? {
      status: 'error';
      reason: Reason;
    }
  : {
      status: 'error';
      reason: Reason;
    } & AdditionalData;

/**
 *
 * DeepURL is an object representation of whatwg.URL.
 *
 * Since DeepLinkParseResult goes through IPC, anything included in it is subject to Structured
 * Clone Algorithm [1]. As such, getters and setters are dropped which means were not able to pass
 * whatwg.URL without casting it to an object.
 *
 * [1] https://developer.mozilla.org/en-US/docs/Web/API/Web_Workers_API/Structured_clone_algorithm
 */
export type DeepURL = {
  /**
   * host is the hostname plus the port.
   */
  host: string;
  /**
   * hostname is the host without the port, e.g. if the host is "example.com:4321", the hostname is
   * "example.com".
   */
  hostname: string;
  port: string;
  /**
   * username is percent-decoded username from the URL. whatwg-url encodes usernames found in URLs.
   * parseDeepLink decodes them so that other parts of the app don't have to deal with this.
   */
  username: string;
  /**
   * pathname is the path from the URL with the leading slash included, e.g. if the URL is
   * "teleport://example.com/connect_my_computer", the pathname is "/connect_my_computer"
   */
  pathname: `/${Path}`;
};

// We're able to get away with defining the path like this only because we don't use matchPath from
// React Router v5 like uri.ts does. Once we get to more complex use cases that will use matchPath,
// we'll likely have to sacrifice some type safety.
export type Path = 'connect_my_computer';

/**
 * parseDeepLink receives a full URL of a deep link passed to Connect, e.g.
 * teleport://foo.example.com:4321/connect_my_computer and returns its parsed form if the underlying
 * URI is supported by the app.
 *
 * Returning a parsed form was a conscious decision – this way it's clear that the parsed form is
 * valid and can be passed along safely from the main process to the renderer vs raw string URLs
 * which don't carry any information by themselves about their validity – in that scenario, they'd
 * have to be parsed on both ends.
 */
export function parseDeepLink(rawUrl: string): DeepLinkParseResult {
  let whatwgURL: whatwg.URL;
  try {
    whatwgURL = new whatwg.URL(rawUrl);
  } catch (error) {
    if (error instanceof TypeError) {
      // Invalid URL.
      return { status: 'error', reason: 'malformed-url', error };
    }
    throw error;
  }

  if (whatwgURL.protocol !== `${CUSTOM_PROTOCOL}:`) {
    return {
      status: 'error',
      reason: 'unknown-protocol',
      protocol: whatwgURL.protocol,
    };
  }

  if (whatwgURL.pathname !== '/connect_my_computer') {
    return { status: 'error', reason: 'unsupported-uri' };
  }

  const { host, hostname, port, username, pathname } = whatwgURL;
  const url: DeepURL = {
    host,
    hostname,
    port,
    // whatwg-url percent-encodes usernames. We decode them here so that the rest of the app doesn't
    // have to do this. https://url.spec.whatwg.org/#set-the-username
    //
    // What's more, Chrome, unlike Firefox and Safari, won't even trigger a custom protocol prompt
    // when clicking on a link that includes a username with an @ symbol that is not
    // percent-encoded, e.g. teleport://alice@example.com@example.com/connect_my_computer.
    // TODO(ravicious): Move this comment to the place that will actually generate deep links in the
    // Web UI.
    username: decodeURIComponent(username),
    pathname,
  };

  return { status: 'success', url };
}
