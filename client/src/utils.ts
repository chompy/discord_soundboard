/** LOGGING **/
export const log = (message: string) => console.log(`> ${message}`);
export const error = (message: string) => console.error(`> ERROR: ${message}`);

export const isNotAuthenticatedError = (error: unknown) =>
    (typeof error === 'string' && error.includes('not authenticated')) ||
    (error instanceof Error && error.message.includes('not authenticated'));

export const handleNotAuthenticatedError = (err: unknown) => {
    if (isNotAuthenticatedError(err)) {
        window.location.href = '/login';
    }
};
