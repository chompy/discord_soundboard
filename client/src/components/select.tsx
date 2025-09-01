import React from 'react';

export type SelectProperties = {
    options: string[];
    onChange?: (index: number) => void;
};

function Select({ options, onChange }: SelectProperties) {
    const selectId = React.useId();
    return (
        <>
            <select onChange={(e) => onChange?.(parseInt(e.target.value))}>
                {options.map((label, index) => (
                    <option key={`${selectId}-${index}`} value={index}>
                        {label}
                    </option>
                ))}
            </select>
        </>
    );
}

export default Select;
