import { observable, WritableObservable } from 'micro-observables';

import { globalStorage } from 'globalStorage';
import { Auth } from 'types';
import { getCookie, setCookie } from 'utils/browser';
import { Loadable, Loaded, NotLoaded } from 'utils/loadable';

export const AUTH_COOKIE_KEY = 'auth';

const clearAuthCookie = (): void => {
  document.cookie = `${AUTH_COOKIE_KEY}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
};

/**
 * set the auth cookie if it's not already set.
 *
 * @param token auth token
 */
const ensureAuthCookieSet = (token: string): void => {
  if (!getCookie(AUTH_COOKIE_KEY)) setCookie(AUTH_COOKIE_KEY, token);
};

const internalAuth = new WritableObservable<Loadable<Auth>>(NotLoaded);
const internalAuthChecked = new WritableObservable(false);

export const auth = internalAuth.readOnly();
export const authChecked = internalAuthChecked.readOnly();

export const reset = (): void => {
  clearAuthCookie();
  globalStorage.removeAuthToken();
  WritableObservable.batch(() => {
    internalAuth.set(NotLoaded);
    internalAuthChecked.set(false);
  });
};
export const setAuth = (newAuth: Auth): void => {
  if (newAuth.token) {
    ensureAuthCookieSet(newAuth.token);
    globalStorage.authToken = newAuth.token;
  }
  internalAuth.set(Loaded(newAuth));
};
export const setAuthChecked = (): void => internalAuthChecked.set(true);
export const selectIsAuthenticated = auth.select((a) =>
  Loadable.match(a, {
    Loaded: (au) => au.isAuthenticated,
    NotLoaded: () => false,
  }),
);

interface AuthState {
  auth: Loadable<Auth>;
  isChecked: boolean;
}

const defaultState: AuthState = {
  auth: NotLoaded,
  isChecked: false,
};

class AuthStore {
  protected state: WritableObservable<AuthState> = observable(defaultState);
  public readonly auth = this.state.select((s) => s.auth);
  public readonly isChecked = this.state.select((s) => s.isChecked);
  public readonly isAuthenticated = this.auth.select((auth) => {
    return Loadable.match(auth, {
      Loaded: (a) => a.isAuthenticated,
      NotLoaded: () => false,
    });
  });

  public setAuth(newAuth: Auth) {
    if (newAuth.token) {
      ensureAuthCookieSet(newAuth.token);
      globalStorage.authToken = newAuth.token;
    }
    this.state.update((s) => ({ ...s, auth: Loaded(newAuth) }));
  }

  public setIsChecked() {
    this.state.update((s) => ({ ...s, isChecked: true }));
  }

  public reset() {
    clearAuthCookie();
    globalStorage.removeAuthToken();
    this.state.set(defaultState);
  }
}

const authStore = new AuthStore();

export default authStore;
