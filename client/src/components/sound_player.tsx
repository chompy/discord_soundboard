import { useCallback, useEffect, useRef, useState } from 'react';
import { Sound } from '../api';
import { SoundList } from '../hooks/sound_list';
import { SoundPlayerSound } from './sound_player_sound';


export type SoundPlayerProperties = {
    soundList: SoundList;
    enableKeyBinding?: boolean;
    onKeybindSelect?: (sound: Sound | null) => void;
};

function SoundPlayer({ soundList, enableKeyBinding, onKeybindSelect }: SoundPlayerProperties) {
    const { isLoading, categories, sounds } = soundList;
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

    if (isLoading) return;

    const favorites = soundList.favorites.get();

    return (
        <div ref={elementRef} className="sound-player">
            {favorites.length > 0 && 
            <div className='favorites'>
                <h2>Favorites</h2>
                <div className='categories'>

                    {favorites.map(({soundId}) => {
                        const sound = soundList.sounds.get().find((sound) => sound.id === soundId);
                        return sound ? <SoundPlayerSound 
                            sound={sound} 
                            soundList={soundList}
                            nameMaxWidth={nameMaxWidth}
                            enableKeyBinding={enableKeyBinding}
                            onKeybindSelect={onKeybindSelect}
                        /> : null;
                    })}

                </div>
            </div>}
            <div className="categories">
                {categories.get().map((category) => (
                    <div key={`category-${category.id}`} className="category">
                        <h2 className="category-name">{category.name}</h2>
                        <div className="sounds">
                            {sounds
                                .get()
                                .filter(
                                    (sound) => sound.categoryId === category.id
                                )
                                .map((sound) => <SoundPlayerSound 
                                    sound={sound}
                                    soundList={soundList}
                                    nameMaxWidth={nameMaxWidth}
                                    enableKeyBinding={enableKeyBinding}
                                    onKeybindSelect={onKeybindSelect}
                                />)
                            }
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}

export default SoundPlayer;
