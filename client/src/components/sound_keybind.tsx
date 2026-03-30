import { useCallback, useEffect, useMemo, useState } from "react";
import { Sound, api } from "../api";
import { SoundList } from "../hooks/sound_list";

const availableKeybindKeys = [
    "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", 
    "p", "q", "r", "t", "u", "v", "w", "x", "y", "z", "1", "2", "3", "4", 
    "5", "6", "7", "8", "9", "0"
];

export type SoundKeybindProperties = {
    sound: Sound;
    soundList: SoundList;
    enabled?: boolean;
    onSelect?: (sound: Sound | null) => void;
};

export function SoundKeybind({sound, soundList, enabled, onSelect}: SoundKeybindProperties) {

    const [active, setActive] = useState(false);
    const keybind = soundList.keybinds.get().find((keybind) => keybind.soundId === sound.id)

    useEffect(() => {
        const onKeyDown = (e: KeyboardEvent) => {
            if (!enabled || !active) return;
            if (e.key == 'Delete') {
                soundList.setKeybind(sound, null);
                return;
            }
            if (!keybind || availableKeybindKeys.includes(e.key) && e.key != keybind.key) {
                soundList.setKeybind(sound, e.key);
            }
        };
        window.addEventListener('keydown', onKeyDown);
        return () => {
            window.removeEventListener('keydown', onKeyDown)
        }
    }, [sound, soundList, active, keybind, enabled])

    const onEnter = useCallback(() => {
        if (enabled) {
            setActive(true);
            onSelect?.(sound);
        }
    }, [sound, enabled]);

    const onExit = useCallback(async () => {
        setActive(false);
        onSelect?.(null);
    }, []);


    return <span
        className={`sound-keybind${active ? ' active' : ''}`}
        onMouseEnter={onEnter}
        onMouseLeave={onExit}
        onClick={(e) => e.preventDefault()}
    >
        {keybind ? keybind.key : '-'}
    </span>
    
}