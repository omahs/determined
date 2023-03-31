const exhaustive = (v: never): never => v;

type MatchArgs<T, U> =
  | {
      Loaded: (data: T) => U;
      NotLoaded: () => U;
    }
  | {
      Loaded: (data: T) => U;
      _: () => U;
    }
  | {
      NotLoaded: () => U;
      _: () => U;
    };

class Loadable_<T> {
  _tag: 'Loaded' | 'NotLoaded';
  data: T | undefined;

  constructor(tag: 'Loaded' | 'NotLoaded', data: T | undefined) {
    this._tag = tag;
    this.data = data;
  }

  /**
   * The map() function creates a new Loadable with the result of calling
   * the provided function on the contained value in the passed Loadable.
   *
   * If the passed Loadable is NotLoaded then the return value is NotLoaded
   */
  map<U>(fn: (t: T) => U): Loadable<U> {
    switch (this._tag) {
      case 'Loaded':
        return new Loadable_('Loaded', fn(this.data!)) as Loadable<U>;
      case 'NotLoaded':
        return new Loadable_<U>('NotLoaded', undefined) as Loadable<U>;
      default:
        return exhaustive(this._tag);
    }
  }
  static map<T, U>(l: Loadable<T>, fn: (_: T) => U): Loadable<U> {
    return l.map(fn);
  }

  /**
   * The flatMap() function creates a new Loadable with the result of calling
   * the provided function on the contained value in the passed Loadable and then
   * flattening the result.
   *
   * If any of the passed or returned Loadables is NotLoaded, the result is NotLoaded.
   */
  flatMap<U>(fn: (_: T) => Loadable<U>): Loadable<U> {
    switch (this._tag) {
      case 'Loaded':
        return fn(this.data!) as Loadable<U>;
      case 'NotLoaded':
        return new Loadable_<U>('NotLoaded', undefined) as Loadable<U>;
      default:
        return exhaustive(this._tag);
    }
  }
  static flatMap<T, U>(l: Loadable<T>, fn: (_: T) => Loadable<U>): Loadable<U> {
    return l.flatMap(fn);
  }

  /**
   * Performs a side-effecting function if the passed Loadable is Loaded.
   */
  forEach<U>(fn: (_: T) => U): void {
    switch (this._tag) {
      case 'Loaded': {
        fn(this.data!);
        return;
      }
      case 'NotLoaded':
        return;
      default:
        exhaustive(this._tag);
    }
  }
  static forEach<T, U>(l: Loadable<T>, fn: (_: T) => U): void {
    return l.forEach(fn);
  }

  /**
   * If the passed Loadable is Loaded this returns the data, otherwise
   * it returns the default value.
   */
  getOrElse(def: T): T {
    switch (this._tag) {
      case 'Loaded':
        return this.data!;
      case 'NotLoaded':
        return def;
      default:
        return exhaustive(this._tag);
    }
  }
  static getOrElse<T>(def: T, l: Loadable<T>): T {
    return l.getOrElse(def);
  }

  /**
   * Allows you to match out the cases in the Loadable with named
   * arguments.
   */
  match<U>(cases: MatchArgs<T, U>): U {
    switch (this._tag) {
      case 'Loaded':
        return 'Loaded' in cases ? cases.Loaded(this.data!) : cases._();
      case 'NotLoaded':
        return 'NotLoaded' in cases ? cases.NotLoaded() : cases._();
      default:
        return exhaustive(this._tag);
    }
  }
  static match<T, U>(l: Loadable<T>, cases: MatchArgs<T, U>): U {
    return l.match(cases);
  }

  /** Like `match` but without argument names */
  quickMatch<U>(def: U, f: (data: T) => U): U {
    switch (this._tag) {
      case 'Loaded':
        return f(this.data!);
      case 'NotLoaded':
        return def;
      default:
        return exhaustive(this._tag);
    }
  }
  static quickMatch<T, U>(l: Loadable<T>, def: U, f: (data: T) => U): U {
    return l.quickMatch(def, f);
  }

  /**
   * Groups up all passed Loadables. NotFound takes priority over
   * NotLoaded so all([NotLoaded, NotFound, Loaded(4)]) returns NotFound
   */
  static all<A>(ls: [Loadable<A>]): Loadable<[A]>;
  static all<A, B>(ls: [Loadable<A>, Loadable<B>]): Loadable<[A, B]>;
  static all<A, B, C>(ls: [Loadable<A>, Loadable<B>, Loadable<C>]): Loadable<[A, B, C]>;
  static all<A, B, C, D>(
    ls: [Loadable<A>, Loadable<B>, Loadable<C>, Loadable<D>],
  ): Loadable<[A, B, C, D]>;
  static all<A, B, C, D, E>(
    ls: [Loadable<A>, Loadable<B>, Loadable<C>, Loadable<D>, Loadable<E>],
  ): Loadable<[A, B, C, D, E]>;
  static all(ls: Array<Loadable<unknown>>): Loadable<Array<unknown>> {
    const res: unknown[] = [];
    for (const l of ls) {
      if (l._tag === 'NotLoaded') {
        return new Loadable_<unknown[]>('NotLoaded', undefined) as Loadable<unknown[]>;
      }
      res.push(l.data);
    }
    return new Loadable_('Loaded', res) as Loadable<unknown[]>;
  }

  /** Allows you to use Loadables with React's Suspense component */
  waitFor(): T {
    switch (this._tag) {
      case 'Loaded':
        return this.data!;
      case 'NotLoaded':
        throw Promise.resolve(undefined);
      default:
        return exhaustive(this._tag);
    }
  }
  static waitFor<T>(l: Loadable<T>): T {
    return l.waitFor();
  }
  get isLoaded(): boolean {
    return this._tag === 'Loaded';
  }
  static isLoaded<T>(
    l: Loadable<T>,
  ): l is { _tag: 'Loaded'; data: T; isLoaded: true; isNotLoaded: false } & Omit<
    Loadable_<T>,
    'isLoaded' | 'data'
  > {
    return l.isLoaded;
  }
  get isNotLoaded(): boolean {
    return this._tag === 'NotLoaded';
  }
  static isNotLoaded<T>(
    l: Loadable<T>,
  ): l is { _tag: 'NotLoaded'; isLoaded: false; isNotLoaded: true } & Omit<
    Loadable_<T>,
    'isLoaded' | 'data'
  > {
    return l.isNotLoaded;
  }

  /** Returns true if the passed object is a Loadable */
  static isLoadable<T, Z>(l: Loadable<T> | Z): l is Loadable<T> {
    return ['Loaded', 'NotLoaded', 'NotFound'].includes((l as Loadable<T>)?._tag);
  }
}

export type Loadable<T> =
  | ({
      _tag: 'Loaded';
      data: T;
      isLoaded: true;
      isNotLoaded: false;
    } & Omit<Loadable_<T>, '_tag' | 'isLoaded' | 'isNotLoaded' | 'data'>)
  | ({
      _tag: 'NotLoaded';
      isLoaded: false;
      isNotLoaded: true;
    } & Omit<Loadable_<T>, '_tag' | 'isLoaded' | 'isNotLoaded' | 'data'>);

// There's no real way to add methods to a union type in typescript except for with Proxies
// but Proxies don't handle generics correctly. We have to "lie" to typescript here to convince
// it that our class is a union type. It's also impossible to write custom guard types
// as methods on a class so we have to lie to it about the return type of all of our guard methods.
const Loaded = <T>(data: T): Loadable<T> => new Loadable_('Loaded', data) as Loadable<T>;
const NotLoaded: Loadable<never> = new Loadable_('NotLoaded', undefined) as Loadable<never>;

export const Loadable = Loadable_;

export { Loaded, NotLoaded };
