import React, { useState, useEffect, useRef } from 'react';

import UseSoundList from '../hooks/sound_list';
import { api } from '../api';

export type SoundPlayerProperties = {
    guildId: string;
};

function SoundPlayer({ guildId }: SoundPlayerProperties) {
    const { isLoading, categories, sounds } = UseSoundList(guildId);

    if (isLoading) return;

    return (
        <div className="sound-player">
            <div className="categories">
                {categories.map((category) => (
                    <div key={`category-${category.id}`} className="category">
                        <div className="category-name">{category.name}</div>
                        <div className="sounds">
                            {sounds
                                .filter(
                                    (sound) => sound.categoryId === category.id
                                )
                                .map((sound) => (
                                    <div
                                        key={`sound-${sound.id}`}
                                        className="sound"
                                    >
                                        <span className="sound-name">
                                            {sound.name}
                                        </span>
                                        <span className="sound-options">
                                            <a
                                                href="#"
                                                onClick={(e) => {
                                                    e.preventDefault();
                                                    api.playSound(sound);
                                                }}
                                            >
                                                Play
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
