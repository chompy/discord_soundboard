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

/** KEYBINDS */
const keybindStorageKey = 'chompy_soundboard_keybinds'
const availableKeybindKeys = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '0'];

export const getSoundKeybinds = () => {
    const data = window.localStorage.getItem(keybindStorageKey)
    return (data ? JSON.parse(data) : {}) as Record<string, number>;
}

export const setSoundKeybind = (key: string, soundId: number) => {
    if (availableKeybindKeys.includes(key)) {
        log(`Bind sound ${soundId} to ${key}`)
        const data = getSoundKeybinds();
        // unset other uses of sound
        for (let key in data) {
            if (data[key] === soundId) {
                delete data[key];
            }
        }
        // set sound
        data[key] = soundId;
        window.localStorage.setItem(keybindStorageKey, JSON.stringify(data))
    }
}

export const getSoundKeybindForKey = (key: string) => {
    const data = getSoundKeybinds();
    return key in data ? data[key] : null;
}