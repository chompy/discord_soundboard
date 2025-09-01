import React from 'react';

export type ButtonProperties = {
    label: string;
    disabled?: boolean;
    onClick?: () => void;
};

function Button({ label, disabled, onClick }: ButtonProperties) {
    return (
        <button disabled={disabled} className="pure-button" onClick={onClick}>
            {label}
        </button>
    );
}

export default Button;
