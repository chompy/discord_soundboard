import { SoundList } from '../hooks/sound_list';
import { Sound, api } from '../api';
import { useCallback, useEffect, useRef, useState } from 'react';
import { getSoundKeybinds, setSoundKeybind } from '../utils';

export type SoundPlayerProperties = {
    soundList: SoundList;
    enableKeyBinding?: boolean;
    onSelect?: (sound: Sound | null) => void;
};

function SoundPlayer({ soundList, enableKeyBinding, onSelect }: SoundPlayerProperties) {
    const { isLoading, categories, sounds } = soundList;
    const [selectedSound, setSelectedSound] = useState<Sound | null>(null);
    const [nameMaxWidth, setNameMaxWidth] = useState(250);

    // resize name width
    const elementRef = useRef<HTMLDivElement>(null);
    const resize = useCallback(() => {
        if (!elementRef.current) return;
        const soundElement = elementRef.current.getElementsByClassName('sound')[0];
        if (!soundElement) return;
        const soundOptionsElement = soundElement.getElementsByClassName('sound-options')[0] as HTMLSpanElement;
        setNameMaxWidth(soundElement.clientWidth - soundOptionsElement.clientWidth - 30);
    }, [elementRef.current])
    useEffect(() => {
        resize();
        window.addEventListener('resize', resize);
        return () => {
            window.removeEventListener('resize', resize);
        }
    }, [elementRef.current]);

    // set sound keybinds
    const onKeyDown = useCallback((e: KeyboardEvent) => {
        enableKeyBinding && selectedSound && setSoundKeybind(e.key, selectedSound.id);
        setSelectedSound(null);
    }, [selectedSound, enableKeyBinding])

    useEffect(() => {
        window.addEventListener('keydown', onKeyDown);
        return () => {
            window.removeEventListener('keydown', onKeyDown)
        }
    }, [selectedSound])
    const soundKeybinds = Object.entries(getSoundKeybinds());
    const findSoundKeybind = (sound: Sound) => {
        const res = soundKeybinds.find(([key, soundId]) => sound.id === soundId);
        return res ? res[0] : null;
    }

    if (isLoading) return;
    return (
        <div ref={elementRef} className="sound-player">
            <div className="categories">
                {categories.get().map((category) => (
                    <div key={`category-${category.id}`} className="category">
                        <div className="category-name">{category.name}</div>
                        <div className="sounds">
                            {sounds
                                .get()
                                .filter(
                                    (sound) => sound.categoryId === category.id
                                )
                                .map((sound) => (
                                    <div
                                        key={`sound-${sound.id}`}
                                        className="sound"
                                    >
                                        <span className="sound-name" title={sound.name} style={{maxWidth: `${nameMaxWidth}px`}}>
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
                                                className={`sound-keybind${selectedSound && selectedSound.id === sound.id ? ' selected' : ''}`}
                                                href="#"
                                                data-sound-id={sound.id}
                                                onMouseEnter={() => {setSelectedSound(sound); onSelect?.(sound)}}
                                                onMouseLeave={() => {setSelectedSound(null); onSelect?.(null)}}
                                                onClick={(e) => e.preventDefault()}
                                            >
                                                {findSoundKeybind(sound) ?? '-'}
                                            </a>
                                        </span>
                                    </div>
                                ))}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}

export default SoundPlayer;
