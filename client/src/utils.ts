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
const availableKeybindKeys = [
    "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", 
    "p", "q", "r", "t", "u", "v", "w", "x", "y", "z", "1", "2", "3", "4", 
    "5", "6", "7", "8", "9", "0"
];

export const getSoundKeybinds = () => {
    const data = window.localStorage.getItem(keybindStorageKey)
    return (data ? JSON.parse(data) : {}) as Record<string, number>;
}

export const setSoundKeybind = (key: string, soundId: number): boolean => {
    if (!availableKeybindKeys.includes(key) && key !== 'Delete') return false;

    // fetch keybinds
    const data = getSoundKeybinds();

    // unset other uses of sound
    for (let key in data) {
        if (data[key] === soundId) {
            log(`Unbind sound ${soundId} from ${key}`)
            delete data[key];
        }
    }

    // set sound
    if (key !== 'Delete') {
        log(`Bind sound ${soundId} to ${key}`)
        data[key] = soundId;
    }

    // save changes
    window.localStorage.setItem(keybindStorageKey, JSON.stringify(data))
    return true;
}

export const getSoundKeybindForKey = (key: string) => {
    const data = getSoundKeybinds();
    return key in data ? data[key] : null;
}