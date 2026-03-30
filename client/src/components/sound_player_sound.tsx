import { useCallback, useMemo, useState } from "react";
import { Sound, api } from "../api";
import { SoundKeybind } from "./sound_keybind";
import { SoundList } from "../hooks/sound_list";



export type SoundPlayerSoundProperties = {
    sound: Sound;
    soundList: SoundList;
    nameMaxWidth?: number;
    enableKeyBinding?: boolean;
    onKeybindSelect?: (sound: Sound | null) => void;
};

export function SoundPlayerSound({sound, soundList, nameMaxWidth, enableKeyBinding, onKeybindSelect}: SoundPlayerSoundProperties) {

    const isFavorite = useMemo(() => soundList.favorites.get().find((fav) => fav.soundId === sound.id), [soundList]);

    return <div
        key={`sound-${sound.id}`}
        className={`sound${isFavorite ? ' favorite' : ''}`}
    >
        <span className="sound-name" title={sound.name} style={{maxWidth: nameMaxWidth ? `${nameMaxWidth}px` : ''}}>
            {sound.name}
        </span>
        <span className="sound-options">
            <a
                className='pure-button'
                href="#"
                onClick={(e) => {
                    e.preventDefault();
                    api.playSound(sound);
                }}
            >
                Play
            </a>
            <a
                className='favorite-btn'
                href="#"
                onClick={(e) => {
                    e.preventDefault();
                    soundList.setFavorite(sound, !isFavorite);
                }}
            >
                ⭐
            </a>
            <SoundKeybind sound={sound} soundList={soundList} enabled={enableKeyBinding} onSelect={onKeybindSelect} />
        </span>
    </div>
}