export class APIError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly info: unknown,
  ) {
    super(message);
    this.name = 'APIError';
  }
}

export const customFetch = async <T>(
  url: string,
  options: RequestInit,
): Promise<T> => {
  const response = await fetch(url, { ...options, credentials: 'include' });
  const text = [204, 205, 304].includes(response.status)
    ? ''
    : await response.text();
  const data = text ? (JSON.parse(text) as unknown) : {};
  if (!response.ok) {
    const problem = data as { detail?: string; title?: string };
    throw new APIError(
      problem.detail ?? problem.title ?? `Request failed (${response.status})`,
      response.status,
      data,
    );
  }
  return data as T;
};
